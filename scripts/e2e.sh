#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_CONTAINER="${PG_CONTAINER:-ks-postgres}"

command -v grpcurl >/dev/null 2>&1 || { echo "缺少 grpcurl，请先安装：go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "缺少 python3"; exit 1; }

json_field() {
  python3 -c 'import json,sys; print(json.load(sys.stdin)['"$1"'])'
}

grpc_user() {
  grpcurl -plaintext \
    -import-path "$ROOT_DIR/service/user/api" \
    -import-path "$ROOT_DIR/service/user/third_party" \
    -proto user/v1/user.proto "$@"
}
grpc_cart() {
  grpcurl -plaintext \
    -import-path "$ROOT_DIR/service/cart/api" \
    -import-path "$ROOT_DIR/service/cart/third_party" \
    -proto cart/v1/cart.proto "$@"
}
grpc_order() {
  grpcurl -plaintext \
    -import-path "$ROOT_DIR/service/order/api" \
    -import-path "$ROOT_DIR/service/order/third_party" \
    -proto order/v1/order.proto "$@"
}
grpc_payment() {
  grpcurl -plaintext \
    -import-path "$ROOT_DIR/service/payment/api" \
    -proto payment/v1/payment.proto "$@"
}

psqlq() {
  docker exec -i "$PG_CONTAINER" psql -U postgres -d "$1" -tAc "$2"
}

E2E_USER=$((900000 + RANDOM % 99999))
E2E_MOBILE="138$(printf '%08d' $((RANDOM % 100000000)))"
echo "==> E2E user: $E2E_USER / $E2E_MOBILE"

INV_BEFORE=$(psqlq shop_inventory "SELECT inventory FROM inventories WHERE sku_id=1")
SOLD_BEFORE=$(psqlq shop_goods "SELECT sold_num FROM goods WHERE id=1")

echo "==> 创建测试用户"
USER_JSON=$(grpc_user -d "{\"mobile\":\"$E2E_MOBILE\",\"password\":\"12345678\",\"nickName\":\"E2E\"}" \
  127.0.0.1:50051 user.v1.User/CreateUser)
E2E_USER=$(echo "$USER_JSON" | json_field '"id"')
echo "user_id=$E2E_USER"

echo "==> 创建收货地址"
ADDR_JSON=$(grpc_user -d "{\"uid\":$E2E_USER,\"name\":\"E2E 测试\",\"mobile\":\"$E2E_MOBILE\",\"Province\":\"广东省\",\"City\":\"深圳市\",\"Districts\":\"南山区\",\"address\":\"测试路 1 号\",\"post_code\":\"518000\",\"is_default\":1}" \
  127.0.0.1:50051 user.v1.User/CreateAddress)
ADDR_ID=$(echo "$ADDR_JSON" | json_field '"id"')
echo "address_id=$ADDR_ID"

echo "==> 添加购物车"
CART_JSON=$(grpc_cart -d "{\"userId\":$E2E_USER,\"goodsId\":1,\"goodsSn\":\"E2E-SN-001\",\"goodsName\":\"E2E iPhone\",\"skuId\":1,\"goodsPrice\":849900,\"goodsNum\":2,\"isSelect\":true}" \
  127.0.0.1:50053 cart.v1.Cart/CreateCart)
CART_ID=$(echo "$CART_JSON" | json_field '"id"')
echo "cart_id=$CART_ID"

echo "==> 创建订单"
ORDER_JSON=$(grpc_order -d "{\"userId\":$E2E_USER,\"address\":$ADDR_ID,\"cartItem\":[{\"cartId\":$CART_ID,\"skuId\":1,\"skuNum\":2}]}" \
  127.0.0.1:50054 order.v1.Order/CreateOrder)
ORDER_SN=$(echo "$ORDER_JSON" | json_field '"orderSn"')
ORDER_TOTAL=$(echo "$ORDER_JSON" | json_field '"total"')
echo "order_sn=$ORDER_SN total=$ORDER_TOTAL"

echo "==> 创建支付单并模拟支付成功"
PAY_JSON=$(grpc_payment -d "{\"userId\":$E2E_USER,\"orderSn\":\"$ORDER_SN\",\"amount\":$ORDER_TOTAL,\"channel\":\"mock\"}" \
  127.0.0.1:50056 payment.v1.Payment/CreatePayment)
PAY_NO=$(echo "$PAY_JSON" | json_field '"paymentNo"')
grpc_payment -d "{\"paymentNo\":\"$PAY_NO\",\"tradeNo\":\"E2E-$(date +%s)\",\"success\":true}" \
  127.0.0.1:50056 payment.v1.Payment/PaymentCallback >/dev/null

echo "==> 等待事件消费"
sleep 4

echo "==> 校验主链路"
ORDER_STATUS=$(psqlq shop_order "SELECT order_status FROM orders WHERE order_sn='$ORDER_SN'")
PAY_STATUS=$(psqlq shop_payment "SELECT status FROM payments WHERE payment_no='$PAY_NO'")
INV_AFTER=$(psqlq shop_inventory "SELECT inventory FROM inventories WHERE sku_id=1")
SOLD_AFTER=$(psqlq shop_goods "SELECT sold_num FROM goods WHERE id=1")
PAID_COUNT=$(psqlq shop_order "SELECT count(*) FROM order_event_outbox WHERE order_sn='$ORDER_SN' AND event_type='order.paid' AND status=1")

[[ "$ORDER_STATUS" == "2" ]] || { echo "FAIL: order status=$ORDER_STATUS"; exit 1; }
[[ "$PAY_STATUS" == "2" ]] || { echo "FAIL: payment status=$PAY_STATUS"; exit 1; }
[[ "$INV_AFTER" == "$((INV_BEFORE - 2))" ]] || { echo "FAIL: inventory $INV_BEFORE -> $INV_AFTER"; exit 1; }
[[ "$SOLD_AFTER" == "$((SOLD_BEFORE + 2))" ]] || { echo "FAIL: sold $SOLD_BEFORE -> $SOLD_AFTER"; exit 1; }
[[ "$PAID_COUNT" == "1" ]] || { echo "FAIL: order.paid count=$PAID_COUNT"; exit 1; }
echo "PASS: 下单/支付/扣库存/加销量 均正确（库存 ${INV_BEFORE}->${INV_AFTER}，销量 ${SOLD_BEFORE}->${SOLD_AFTER}）"

echo "==> 校验取消释放"
CART2_JSON=$(grpc_cart -d "{\"userId\":$E2E_USER,\"goodsId\":3,\"goodsSn\":\"E2E-SN-002\",\"goodsName\":\"E2E Lenovo\",\"skuId\":4,\"goodsPrice\":529900,\"goodsNum\":1,\"isSelect\":true}" \
  127.0.0.1:50053 cart.v1.Cart/CreateCart)
CART2_ID=$(echo "$CART2_JSON" | json_field '"id"')
ORDER2_JSON=$(grpc_order -d "{\"userId\":$E2E_USER,\"address\":$ADDR_ID,\"cartItem\":[{\"cartId\":$CART2_ID,\"skuId\":4,\"skuNum\":1}]}" \
  127.0.0.1:50054 order.v1.Order/CreateOrder)
ORDER2_SN=$(echo "$ORDER2_JSON" | json_field '"orderSn"')
echo "order2_sn=$ORDER2_SN"
sleep 5
grpc_order -d "{\"userId\":$E2E_USER,\"orderSn\":\"$ORDER2_SN\"}" \
  127.0.0.1:50054 order.v1.Order/CancelOrder >/dev/null
sleep 5

ORDER2_STATUS=$(psqlq shop_order "SELECT order_status FROM orders WHERE order_sn='$ORDER2_SN'")
LOCK2_STATUS=$(psqlq shop_inventory "SELECT status FROM inventory_locks WHERE order_sn='$ORDER2_SN'")
echo "order2_status=$ORDER2_STATUS lock2_status=$LOCK2_STATUS"
[[ "$ORDER2_STATUS" == "5" ]] || { echo "FAIL: cancel order status=$ORDER2_STATUS"; exit 1; }
[[ "$LOCK2_STATUS" == "3" ]] || { echo "FAIL: inventory lock status=$LOCK2_STATUS"; exit 1; }
echo "PASS: 取消订单/释放库存 正确"

echo "==> E2E 全部通过"

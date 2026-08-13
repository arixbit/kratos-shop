import {
  App,
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd';
import { Access, useAccess } from '@umijs/max';
import { useEffect, useState } from 'react';
import {
  adminOrderDetail,
  adminOrderList,
  adminRefundOrder,
  adminShipOrder,
} from '@/services/ant-design-pro/api';

const STATUS_MAP: Record<number, { text: string; color: string }> = {
  1: { text: '待支付', color: 'orange' },
  2: { text: '已支付', color: 'blue' },
  3: { text: '已发货', color: 'geekblue' },
  4: { text: '已签收', color: 'cyan' },
  5: { text: '已取消', color: 'default' },
  6: { text: '交易完成', color: 'green' },
  7: { text: '已退款', color: 'red' },
};

const money = (value?: number | string) => `¥${(Number(value || 0) / 100).toFixed(2)}`;

const statusText = (text?: string) => {
  const key = Object.keys(STATUS_MAP).find(
    (k) => STATUS_MAP[Number(k)].text === text,
  );
  return STATUS_MAP[Number(key)] || { text: text || '未知', color: 'default' };
};

export default function Orders() {
  const { message, modal } = App.useApp();
  const access = useAccess();
  const [data, setData] = useState<API.OrderInfo[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [status, setStatus] = useState<number | undefined>();
  const [loading, setLoading] = useState(false);
  const [detail, setDetail] = useState<API.OrderInfo | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [shipRecord, setShipRecord] = useState<API.OrderInfo | null>(null);
  const [shipForm] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      const resp = await adminOrderList({ page, pageSize, status });
      setData(resp.list || []);
      setTotal(resp.total || 0);
    } catch (err) {
      message.error((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, pageSize, status]);

  const openDetail = async (record: API.OrderInfo) => {
    try {
      const resp = await adminOrderDetail(record.orderSn || '');
      setDetail(resp);
      setDetailOpen(true);
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const doShip = async () => {
    const values = await shipForm.validateFields();
    try {
      await adminShipOrder(shipRecord?.orderSn || '', values.post);
      message.success('发货成功');
      setShipRecord(null);
      shipForm.resetFields();
      load();
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const doRefund = (record: API.OrderInfo) => {
    modal.confirm({
      title: '确认退款',
      content: `确定对订单 ${record.orderSn} 执行退款吗？`,
      onOk: async () => {
        try {
          await adminRefundOrder(record.orderSn || '');
          message.success('退款成功');
          load();
        } catch (err) {
          message.error((err as Error).message);
        }
      },
    });
  };

  const columns = [
    { title: '订单号', dataIndex: 'orderSn', width: 200 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (value: string) => {
        const item = statusText(value);
        return <Tag color={item.color}>{item.text}</Tag>;
      },
    },
    { title: '金额', dataIndex: 'total', width: 120, render: money },
    { title: '收货人', dataIndex: 'name', width: 120 },
    { title: '手机号', dataIndex: 'mobile', width: 140 },
    { title: '收货地址', dataIndex: 'address', ellipsis: true },
    { title: '物流单号', dataIndex: 'post', width: 140 },
    { title: '下单时间', dataIndex: 'addTime', width: 170 },
    {
      title: '操作',
      width: 180,
      render: (_: unknown, record: API.OrderInfo) => (
        <Space>
          <Button type="link" size="small" onClick={() => openDetail(record)}>
            详情
          </Button>
          {record.status === '已支付' && (
            <Access accessible={access.can('order:ship')}>
              <Button type="link" size="small" onClick={() => setShipRecord(record)}>
                发货
              </Button>
            </Access>
          )}
          {['已支付', '已发货', '已签收'].includes(record.status || '') && (
            <Access accessible={access.can('order:refund')}>
              <Popconfirm title="确认退款？" onConfirm={() => doRefund(record)}>
                <Button type="link" size="small" danger>
                  退款
                </Button>
              </Popconfirm>
            </Access>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card
        title="订单管理"
        extra={
          <Select
            allowClear
            placeholder="按状态筛选"
            style={{ width: 160 }}
            value={status}
            onChange={setStatus}
            options={Object.entries(STATUS_MAP).map(([value, item]) => ({
              value: Number(value),
              label: item.text,
            }))}
          />
        }
      >
        <Table
          rowKey="orderSn"
          loading={loading}
          columns={columns}
          dataSource={data}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            onChange: (nextPage, nextPageSize) => {
              setPage(nextPage);
              setPageSize(nextPageSize);
            },
          }}
        />
      </Card>

      <Modal
        title="订单详情"
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={null}
        width={720}
      >
        {detail && (
          <>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="订单号">{detail.orderSn}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={statusText(detail.status).color}>{detail.status}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="金额">{money(detail.total)}</Descriptions.Item>
              <Descriptions.Item label="物流单号">{detail.post || '-'}</Descriptions.Item>
              <Descriptions.Item label="收货人">{detail.name}</Descriptions.Item>
              <Descriptions.Item label="手机号">{detail.mobile}</Descriptions.Item>
              <Descriptions.Item label="收货地址" span={2}>
                {detail.address}
              </Descriptions.Item>
              <Descriptions.Item label="下单时间" span={2}>
                {detail.addTime}
              </Descriptions.Item>
            </Descriptions>
            <Typography.Title level={5} style={{ marginTop: 16 }}>
              商品明细
            </Typography.Title>
            <Table
              rowKey="id"
              size="small"
              pagination={false}
              dataSource={detail.goods || []}
              columns={[
                { title: 'SKU ID', dataIndex: 'skuId', width: 100 },
                { title: '商品', dataIndex: 'skuName' },
                { title: '单价', dataIndex: 'skuPrice', width: 120, render: money },
                { title: '数量', dataIndex: 'num', width: 80 },
                { title: '小计', dataIndex: 'totalPrice', width: 120, render: money },
              ]}
            />
          </>
        )}
      </Modal>

      <Modal
        title="订单发货"
        open={!!shipRecord}
        onCancel={() => {
          setShipRecord(null);
          shipForm.resetFields();
        }}
        onOk={doShip}
      >
        <Form form={shipForm} layout="vertical">
          <Form.Item label="订单号">{shipRecord?.orderSn}</Form.Item>
          <Form.Item
            name="post"
            label="物流单号"
            rules={[{ required: true, message: '请输入物流单号' }]}
          >
            <Input placeholder="例如 SF1234567890" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

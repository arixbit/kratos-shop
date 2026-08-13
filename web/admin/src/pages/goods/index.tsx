import {
  App,
  Button,
  Card,
  Descriptions,
  Form,
  Image,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
} from 'antd';
import { Access, useAccess } from '@umijs/max';
import { useEffect, useState } from 'react';
import {
  adminCreateGoods,
  adminDeleteGoods,
  adminGoodsBrands,
  adminGoodsCategories,
  adminGoodsDetail,
  adminGoodsList,
  adminUpdateGoodsStatus,
} from '@/services/ant-design-pro/api';

const money = (value?: number | string) => `¥${(Number(value || 0) / 100).toFixed(2)}`;

export default function Goods() {
  const { message } = App.useApp();
  const access = useAccess();
  const [data, setData] = useState<API.GoodsItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<API.CategoryItem[]>([]);
  const [brands, setBrands] = useState<API.BrandItem[]>([]);
  const [searchForm] = Form.useForm();
  const [detail, setDetail] = useState<API.GoodsDetail | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [createForm] = Form.useForm();

  const load = async (nextPage = page, nextPageSize = pageSize) => {
    setLoading(true);
    try {
      const values = await searchForm.validateFields().catch(() => ({}));
      const resp = await adminGoodsList({
        page: nextPage,
        pageSize: nextPageSize,
        keywords: values.keywords,
        categoryId: values.categoryId,
        brandId: values.brandId,
        skuCode: values.skuCode,
      });
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
  }, [page, pageSize]);

  useEffect(() => {
    adminGoodsCategories().then((resp) => setCategories(resp.list || [])).catch(() => {});
    adminGoodsBrands().then((resp) => setBrands(resp.list || [])).catch(() => {});
  }, []);

  const openDetail = async (record: API.GoodsItem) => {
    try {
      const resp = await adminGoodsDetail(Number(record.id));
      setDetail(resp);
      setDetailOpen(true);
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const toggleStatus = async (record: API.GoodsItem, onSale: boolean) => {
    try {
      await adminUpdateGoodsStatus(Number(record.id), onSale);
      message.success(onSale ? '已上架' : '已下架');
      load();
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const removeGoods = async (record: API.GoodsItem) => {
    try {
      await adminDeleteGoods(Number(record.id));
      message.success('删除成功');
      load();
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const createGoods = async () => {
    const values = await createForm.validateFields();
    try {
      await adminCreateGoods({
        ...values,
        marketPrice: Number(values.marketPrice),
        inventory: Number(values.inventory || 0),
        skuPrice: Number(values.skuPrice),
        promotionPrice: Number(values.promotionPrice || 0),
        skuInventory: Number(values.skuInventory || 0),
      });
      message.success('商品创建成功');
      setCreateOpen(false);
      createForm.resetFields();
      load();
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    {
      title: '商品',
      dataIndex: 'name',
      render: (_: unknown, record: API.GoodsItem) => (
        <>
          <div>{record.name}</div>
          <div style={{ color: '#999', fontSize: 12 }}>{record.goodsSn}</div>
        </>
      ),
    },
    { title: '分类', dataIndex: 'categoryName', width: 100 },
    { title: '品牌', dataIndex: 'brandName', width: 100 },
    { title: 'SKU 数', dataIndex: 'skuCount', width: 90 },
    { title: '价格', dataIndex: 'marketPrice', width: 120, render: money },
    { title: '销量', dataIndex: 'soldNum', width: 90 },
    {
      title: '状态',
      dataIndex: 'onSale',
      width: 90,
      render: (value: boolean) => (
        <Tag color={value ? 'green' : 'red'}>{value ? '在售' : '下架'}</Tag>
      ),
    },
    {
      title: '操作',
      width: 220,
      render: (_: unknown, record: API.GoodsItem) => (
        <Space>
          <Button type="link" size="small" onClick={() => openDetail(record)}>
            详情
          </Button>
          <Access accessible={access.can('goods:status')}>
            <Switch
              size="small"
              checked={record.onSale}
              checkedChildren="上架"
              unCheckedChildren="下架"
              onChange={(checked) => toggleStatus(record, checked)}
            />
          </Access>
          <Access accessible={access.can('goods:delete')}>
            <Popconfirm title="确认删除该商品？" onConfirm={() => removeGoods(record)}>
              <Button type="link" danger size="small">
                删除
              </Button>
            </Popconfirm>
          </Access>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card
        title="商品管理"
        extra={
          <Access accessible={access.can('goods:create')}>
            <Button type="primary" onClick={() => setCreateOpen(true)}>
              添加商品
            </Button>
          </Access>
        }
      >
        <Form form={searchForm} layout="inline" style={{ marginBottom: 16 }}>
          <Form.Item name="keywords">
            <Input placeholder="商品名称 / 编号" allowClear style={{ width: 180 }} />
          </Form.Item>
          <Form.Item name="categoryId">
            <Select
              placeholder="按分类"
              allowClear
              showSearch
              optionFilterProp="label"
              style={{ width: 160 }}
              options={categories.map((item) => ({
                value: Number(item.id),
                label: `${'　'.repeat(Math.max(0, Number(item.level || 1) - 1))}${item.name}`,
              }))}
            />
          </Form.Item>
          <Form.Item name="brandId">
            <Select
              placeholder="按品牌"
              allowClear
              showSearch
              optionFilterProp="label"
              style={{ width: 160 }}
              options={brands.map((item) => ({
                value: Number(item.id),
                label: item.name,
              }))}
            />
          </Form.Item>
          <Form.Item name="skuCode">
            <Input placeholder="SKU 编码" allowClear style={{ width: 160 }} />
          </Form.Item>
          <Space>
            <Button
              type="primary"
              onClick={() => {
                setPage(1);
                load(1);
              }}
            >
              查询
            </Button>
            <Button
              onClick={() => {
                searchForm.resetFields();
                setPage(1);
                load(1);
              }}
            >
              重置
            </Button>
          </Space>
        </Form>
        <Table
          rowKey="id"
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
        title="商品详情"
        open={detailOpen}
        onCancel={() => setDetailOpen(false)}
        footer={null}
        width={860}
      >
        {detail && (
          <>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="商品名称" span={2}>
                {detail.name}
              </Descriptions.Item>
              <Descriptions.Item label="商品编号">{detail.goodsSn}</Descriptions.Item>
              <Descriptions.Item label="分类 / 品牌">
                {detail.categoryName || detail.categoryId} / {detail.brandName || detail.brandId}
              </Descriptions.Item>
              <Descriptions.Item label="市场价">{money(detail.marketPrice)}</Descriptions.Item>
              <Descriptions.Item label="销量">{detail.soldNum}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={detail.onSale ? 'green' : 'red'}>
                  {detail.onSale ? '在售' : '下架'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="新品 / 热卖">
                {detail.isNew ? '是' : '否'} / {detail.isHot ? '是' : '否'}
              </Descriptions.Item>
              <Descriptions.Item label="简介" span={2}>
                {detail.goodsBrief || '-'}
              </Descriptions.Item>
            </Descriptions>

            <h4 style={{ marginTop: 16 }}>SKU 列表</h4>
            <Table
              rowKey="id"
              size="small"
              pagination={false}
              dataSource={detail.skus || []}
              columns={[
                { title: 'SKU 名称', dataIndex: 'skuName' },
                { title: 'SKU 编码', dataIndex: 'skuCode' },
                { title: '价格', dataIndex: 'price', width: 120, render: money },
                {
                  title: '促销价',
                  dataIndex: 'promotionPrice',
                  width: 120,
                  render: (value: number) => (value > 0 ? money(value) : '-'),
                },
                { title: '库存', dataIndex: 'inventory', width: 90 },
                {
                  title: '状态',
                  dataIndex: 'onSale',
                  width: 90,
                  render: (value: boolean) => (
                    <Tag color={value ? 'green' : 'red'}>{value ? '在售' : '下架'}</Tag>
                  ),
                },
              ]}
            />

            {(detail.images || []).length > 0 && (
              <>
                <h4 style={{ marginTop: 16 }}>商品图片</h4>
                <Space wrap>
                  {detail.images?.map((img) => (
                    <Image key={img.url} src={img.url} width={100} height={100} />
                  ))}
                </Space>
              </>
            )}
          </>
        )}
      </Modal>

      <Modal
        title="添加商品"
        open={createOpen}
        onCancel={() => {
          setCreateOpen(false);
          createForm.resetFields();
        }}
        onOk={createGoods}
        width={720}
      >
        <Form form={createForm} layout="vertical">
          <Space size="large" style={{ display: 'flex' }} wrap>
            <Form.Item name="categoryId" label="分类 ID" rules={[{ required: true }]}>
              <InputNumber min={1} style={{ width: 180 }} />
            </Form.Item>
            <Form.Item name="brandId" label="品牌 ID" rules={[{ required: true }]}>
              <InputNumber min={1} style={{ width: 180 }} />
            </Form.Item>
            <Form.Item name="typeId" label="类型 ID" rules={[{ required: true }]}>
              <InputNumber min={1} style={{ width: 180 }} />
            </Form.Item>
            <Form.Item name="name" label="商品名称" rules={[{ required: true }]}>
              <Input style={{ width: 300 }} />
            </Form.Item>
            <Form.Item name="goodsSn" label="商品编号" rules={[{ required: true }]}>
              <Input style={{ width: 300 }} />
            </Form.Item>
            <Form.Item name="marketPrice" label="市场价（分）" rules={[{ required: true }]}>
              <InputNumber min={1} style={{ width: 180 }} />
            </Form.Item>
            <Form.Item name="inventory" label="商品库存">
              <InputNumber min={0} style={{ width: 180 }} />
            </Form.Item>
            <Form.Item name="goodsBrief" label="商品简介">
              <Input style={{ width: 300 }} />
            </Form.Item>
            <Form.Item name="goodsFrontImage" label="主图 URL">
              <Input style={{ width: 300 }} />
            </Form.Item>
          </Space>
          <Space size="large" style={{ display: 'flex' }} wrap>
            <Form.Item name="skuName" label="SKU 名称" rules={[{ required: true }]}>
              <Input style={{ width: 200 }} />
            </Form.Item>
            <Form.Item name="skuCode" label="SKU 编码" rules={[{ required: true }]}>
              <Input style={{ width: 200 }} />
            </Form.Item>
            <Form.Item name="barCode" label="条码" rules={[{ required: true }]}>
              <Input style={{ width: 200 }} />
            </Form.Item>
            <Form.Item name="skuPrice" label="SKU 价格（分）" rules={[{ required: true }]}>
              <InputNumber min={1} style={{ width: 200 }} />
            </Form.Item>
            <Form.Item name="promotionPrice" label="促销价（分）">
              <InputNumber min={0} style={{ width: 200 }} />
            </Form.Item>
            <Form.Item name="skuInventory" label="SKU 库存">
              <InputNumber min={0} style={{ width: 200 }} />
            </Form.Item>
            <Form.Item name="skuImage" label="SKU 图片 URL">
              <Input style={{ width: 260 }} />
            </Form.Item>
          </Space>
          <Space size="large" wrap>
            <Form.Item name="onSale" label="上架" valuePropName="checked" initialValue={true}>
              <Switch />
            </Form.Item>
            <Form.Item name="isNew" label="新品" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item name="isHot" label="热卖" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item name="shipFree" label="包邮" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  );
}

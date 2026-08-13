import {
  App,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
} from 'antd';
import { useEffect, useState } from 'react';
import {
  adminCreateCategory,
  adminDeleteCategory,
  adminGoodsCategories,
  adminUpdateCategory,
} from '@/services/ant-design-pro/api';

export default function Categories() {
  const { message } = App.useApp();
  const [data, setData] = useState<API.CategoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState<API.CategoryItem | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();

  const load = async () => {
    setLoading(true);
    try {
      const resp = await adminGoodsCategories();
      setData(resp.list || []);
    } catch (err) {
      message.error((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    setModalOpen(true);
  };

  const openEdit = (record: API.CategoryItem) => {
    setEditing(record);
    form.setFieldsValue({
      name: record.name,
      parentCategory: Number(record.parentCategory || 0),
      level: Number(record.level || 1),
      sort: 0,
    });
    setModalOpen(true);
  };

  const save = async () => {
    const values = await form.validateFields();
    try {
      if (editing) {
        await adminUpdateCategory({ id: Number(editing.id), ...values });
        message.success('分类更新成功');
      } else {
        await adminCreateCategory(values);
        message.success('分类创建成功');
      }
      setModalOpen(false);
      load();
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const remove = async (record: API.CategoryItem) => {
    try {
      await adminDeleteCategory(Number(record.id));
      message.success('删除成功');
      load();
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const columns = [
    {
      title: '分类名称',
      dataIndex: 'name',
      render: (value: string, record: API.CategoryItem) => (
        <span>
          {'　'.repeat(Math.max(0, Number(record.level || 1) - 1))}
          {value}
          {Number(record.level) === 1 && <Tag color="blue" style={{ marginLeft: 8 }}>一级</Tag>}
        </span>
      ),
    },
    { title: 'ID', dataIndex: 'id', width: 100 },
    { title: '父级 ID', dataIndex: 'parentCategory', width: 100 },
    { title: '层级', dataIndex: 'level', width: 80 },
    {
      title: '操作',
      width: 160,
      render: (_: unknown, record: API.CategoryItem) => (
        <Space>
          <Button type="link" size="small" onClick={() => openEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="确认删除该分类？" onConfirm={() => remove(record)}>
            <Button type="link" danger size="small">
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const parentOptions = data
    .filter((item) => Number(item.level || 1) < 3)
    .map((item) => ({
      value: Number(item.id),
      label: `${'　'.repeat(Math.max(0, Number(item.level || 1) - 1))}${item.name}`,
    }));

  return (
    <div style={{ padding: 24 }}>
      <Card
        title="商品分类"
        extra={
          <Button type="primary" onClick={openCreate}>
            新增分类
          </Button>
        }
      >
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={data}
          pagination={false}
        />
      </Card>

      <Modal
        title={editing ? '编辑分类' : '新增分类'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={save}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="分类名称" rules={[{ required: true, message: '请输入分类名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="parentCategory" label="父级分类" initialValue={0}>
            <Select
              allowClear
              placeholder="不选则为顶级分类"
              options={[{ value: 0, label: '顶级分类' }, ...parentOptions]}
            />
          </Form.Item>
          <Form.Item name="level" label="层级" rules={[{ required: true }]} initialValue={1}>
            <InputNumber min={1} max={3} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="sort" label="排序" initialValue={0}>
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

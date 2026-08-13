import { App, Button, Card, Modal, Popconfirm, Space, Table, Tag, Typography } from 'antd';
import { Access, useAccess } from '@umijs/max';
import { useEffect, useState } from 'react';
import {
  adminUserAddressDelete,
  adminUserAddressList,
  adminUserList,
} from '@/services/ant-design-pro/api';

export default function Users() {
  const { message } = App.useApp();
  const access = useAccess();
  const [data, setData] = useState<API.UserInfo[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [addressUser, setAddressUser] = useState<API.UserInfo | null>(null);
  const [addressList, setAddressList] = useState<API.AddressInfo[]>([]);
  const [addressLoading, setAddressLoading] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const resp = await adminUserList({ page, pageSize });
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

  const openAddress = async (record: API.UserInfo) => {
    setAddressUser(record);
    setAddressLoading(true);
    try {
      const resp = await adminUserAddressList(Number(record.id));
      setAddressList(resp.list || []);
    } catch (err) {
      message.error((err as Error).message);
    } finally {
      setAddressLoading(false);
    }
  };

  const removeAddress = async (record: API.AddressInfo) => {
    try {
      await adminUserAddressDelete(Number(record.id), Number(addressUser?.id));
      message.success('删除成功');
      const resp = await adminUserAddressList(Number(addressUser?.id));
      setAddressList(resp.list || []);
    } catch (err) {
      message.error((err as Error).message);
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 80 },
    { title: '手机号', dataIndex: 'mobile', width: 150 },
    { title: '昵称', dataIndex: 'nickName' },
    { title: '性别', dataIndex: 'gender', width: 100 },
    {
      title: '角色',
      dataIndex: 'role',
      width: 100,
      render: (value: number) => (
        <Tag color={value === 2 ? 'gold' : 'default'}>
          {value === 2 ? '管理员' : '用户'}
        </Tag>
      ),
    },
    {
      title: '操作',
      width: 120,
      render: (_: unknown, record: API.UserInfo) => (
        <Button type="link" size="small" onClick={() => openAddress(record)}>
          地址管理
        </Button>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card title="用户管理">
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
        title={`用户地址：${addressUser?.nickName || addressUser?.mobile || ''}`}
        open={!!addressUser}
        onCancel={() => setAddressUser(null)}
        footer={null}
        width={760}
      >
        <Table
          rowKey="id"
          loading={addressLoading}
          size="small"
          pagination={false}
          dataSource={addressList}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 60 },
            { title: '收货人', dataIndex: 'name', width: 100 },
            { title: '手机号', dataIndex: 'mobile', width: 130 },
            {
              title: '地址',
              render: (_: unknown, record: API.AddressInfo) =>
                [record.Province, record.City, record.Districts, record.address]
                  .filter(Boolean)
                  .join(' '),
            },
            {
              title: '默认',
              dataIndex: 'is_default',
              width: 80,
              render: (value: number) =>
                value === 1 ? <Tag color="green">默认</Tag> : '-',
            },
            {
              title: '操作',
              width: 100,
              render: (_: unknown, record: API.AddressInfo) => (
                <Access accessible={access.can('user:address:delete')}>
                  <Popconfirm title="确认删除该地址？" onConfirm={() => removeAddress(record)}>
                    <Button type="link" danger size="small">
                      删除
                    </Button>
                  </Popconfirm>
                </Access>
              ),
            },
          ]}
        />
        {!addressLoading && addressList.length === 0 && (
          <Typography.Paragraph type="secondary" style={{ textAlign: 'center', marginTop: 16 }}>
            该用户暂无收货地址
          </Typography.Paragraph>
        )}
      </Modal>
    </div>
  );
}

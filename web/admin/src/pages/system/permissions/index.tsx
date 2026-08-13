import { App, Button, Card, Checkbox, Radio, Space, Tag, Typography } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import {
  adminPermissionList,
  adminRolePermissionList,
  adminRolePermissionUpdate,
} from '@/services/ant-design-pro/api';

const ROLES = [
  { value: 1, label: '普通用户' },
  { value: 2, label: '管理员' },
];

export default function Permissions() {
  const { message } = App.useApp();
  const [permissions, setPermissions] = useState<API.PermissionItem[]>([]);
  const [role, setRole] = useState<number>(1);
  const [checked, setChecked] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const loadPermissions = async () => {
    const resp = await adminPermissionList();
    setPermissions(resp.list || []);
  };

  const loadRoleCodes = async (nextRole: number) => {
    setLoading(true);
    try {
      const resp = await adminRolePermissionList(nextRole);
      setChecked(resp.codes || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPermissions().catch((err) => message.error((err as Error).message));
    loadRoleCodes(role).catch((err) => message.error((err as Error).message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const groups = useMemo(() => {
    const map = new Map<string, API.PermissionItem[]>();
    permissions.forEach((item) => {
      const key = item.groupName || '其他';
      if (!map.has(key)) map.set(key, []);
      map.get(key)?.push(item);
    });
    return Array.from(map.entries());
  }, [permissions]);

  const save = async () => {
    setSaving(true);
    try {
      await adminRolePermissionUpdate(role, checked);
      message.success('权限保存成功');
    } catch (err) {
      message.error((err as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <Card
        title="权限管理"
        extra={
          <Button type="primary" loading={saving} onClick={save}>
            保存
          </Button>
        }
      >
        <Typography.Paragraph type="secondary">
          选择角色后勾选权限点，保存后该角色的按钮权限立即生效（前端按钮级隐藏）。
        </Typography.Paragraph>
        <Radio.Group
          value={role}
          onChange={(e) => {
            setRole(e.target.value);
            loadRoleCodes(e.target.value);
          }}
          style={{ marginBottom: 24 }}
        >
          {ROLES.map((item) => (
            <Radio.Button key={item.value} value={item.value}>
              {item.label}
            </Radio.Button>
          ))}
        </Radio.Group>

        <Checkbox.Group value={checked} onChange={(values) => setChecked(values as string[])}>
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            {groups.map(([groupName, items]) => (
              <Card key={groupName} size="small" title={<Tag color="blue">{groupName}</Tag>}>
                <Space wrap>
                  {items.map((item) => (
                    <Checkbox key={item.code} value={item.code} disabled={loading}>
                      {item.name}
                      <Typography.Text type="secondary" style={{ fontSize: 12, marginLeft: 6 }}>
                        {item.code}
                      </Typography.Text>
                    </Checkbox>
                  ))}
                </Space>
              </Card>
            ))}
          </Space>
        </Checkbox.Group>
      </Card>
    </div>
  );
}

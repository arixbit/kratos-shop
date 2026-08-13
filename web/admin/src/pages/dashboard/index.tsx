import { Bar, Column, Pie } from '@ant-design/plots';
import { App, Card, Col, Row, Statistic, Typography } from 'antd';
import { useEffect, useState } from 'react';
import { adminDashboardStats } from '@/services/ant-design-pro/api';

const STATUS_TEXT: Record<number, string> = {
  1: '待支付',
  2: '已支付',
  3: '已发货',
  4: '已签收',
  5: '已取消',
  6: '交易完成',
  7: '已退款',
};

const money = (value?: number | string) => `¥${(Number(value || 0) / 100).toFixed(2)}`;

export default function Dashboard() {
  const { message } = App.useApp();
  const [stats, setStats] = useState<API.DashboardStats | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const resp = await adminDashboardStats();
        setStats(resp);
      } catch (err) {
        message.error((err as Error).message);
      } finally {
        setLoading(false);
      }
    };
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const dailyCount =
    stats?.last30Days?.map((item) => ({
      date: item.date,
      count: Number(item.orderCount || 0),
    })) || [];

  const dailyAmount =
    stats?.last30Days?.map((item) => ({
      date: item.date,
      amount: Number(item.amount || 0) / 100,
    })) || [];

  const statusPie =
    stats?.statusCounts?.map((item) => ({
      type: STATUS_TEXT[Number(item.status)] || `状态${item.status}`,
      value: Number(item.count || 0),
    })) || [];

  const topGoods =
    stats?.topGoods?.map((item) => ({
      name: item.skuName,
      num: Number(item.num || 0),
    })) || [];

  return (
    <div style={{ padding: 24 }}>
      <Typography.Title level={4}>运营看板</Typography.Title>
      <Row gutter={16}>
        <Col flex="1">
          <Card loading={loading}>
            <Statistic title="用户总数" value={Number(stats?.totalUsers || 0)} />
          </Card>
        </Col>
        <Col flex="1">
          <Card loading={loading}>
            <Statistic title="总订单数" value={Number(stats?.totalOrders || 0)} />
          </Card>
        </Col>
        <Col flex="1">
          <Card loading={loading}>
            <Statistic title="总成交额" value={money(stats?.totalSales)} />
          </Card>
        </Col>
        <Col flex="1">
          <Card loading={loading}>
            <Statistic title="今日订单" value={Number(stats?.todayOrders || 0)} />
          </Card>
        </Col>
        <Col flex="1">
          <Card loading={loading}>
            <Statistic title="今日成交额" value={money(stats?.todaySales)} />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="近 30 天订单量" loading={loading}>
            <Column data={dailyCount} xField="date" yField="count" height={260} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="近 30 天成交额（元）" loading={loading}>
            <Column data={dailyAmount} xField="date" yField="amount" height={260} />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="订单状态分布" loading={loading}>
            <Pie
              data={statusPie}
              angleField="value"
              colorField="type"
              innerRadius={0.6}
              height={260}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="热销商品 TOP5（按销量）" loading={loading}>
            <Bar data={topGoods} xField="name" yField="num" height={260} />
          </Card>
        </Col>
      </Row>
    </div>
  );
}

import { useState, useEffect } from 'react';
import { Layout, Menu, Typography, Button, message, Table } from 'antd';
import {
    UserOutlined,
    FileTextOutlined,
    ExclamationCircleOutlined,
    LogoutOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import api from '../utils/api';

const { Header, Sider, Content } = Layout;
const { Title } = Typography;

export default function Dashboard() {
    const navigate = useNavigate();
    const [collapsed, setCollapsed] = useState(false);
    const [currentMenu, setCurrentMenu] = useState('users');
    const [data, setData] = useState([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        fetchData(currentMenu);
    }, [currentMenu]);

    const fetchData = async (menu: string) => {
        setLoading(true);
        try {
            const res: any = await api.get(`/${menu}?page=1&page_size=50`);
            setData(res.items || []);
        } catch {
            setData([]);
        } finally {
            setLoading(false);
        }
    };

    const handleLogout = () => {
        localStorage.removeItem('adminToken');
        message.success('Logged out');
        navigate('/login');
    };

    const renderContent = () => {
        if (currentMenu === 'users') {
            return <Table dataSource={data} rowKey="id" columns={[
                { title: 'ID', dataIndex: 'id' },
                { title: 'Username', dataIndex: 'username' },
                { title: 'Role', dataIndex: 'role' },
                { title: 'Status', dataIndex: 'status' }
            ]} loading={loading} />;
        }
        if (currentMenu === 'papers') {
            return <Table dataSource={data} rowKey="id" columns={[
                { title: 'ID', dataIndex: 'id' },
                { title: 'Title', dataIndex: 'title' },
                { title: 'Zone', dataIndex: 'zone' },
                { title: 'Shit Score', dataIndex: 'shit_score' },
                { title: 'Status', dataIndex: 'status' }
            ]} loading={loading} />;
        }
        if (currentMenu === 'flags') {
            return <Table dataSource={data} rowKey="id" columns={[
                { title: 'ID', dataIndex: 'id' },
                { title: 'Target ID', dataIndex: 'target_id' },
                { title: 'Reason', dataIndex: 'reason' },
                { title: 'Status', dataIndex: 'status' }
            ]} loading={loading} />;
        }
        return <div>Select a menu</div>;
    };

    return (
        <Layout style={{ minHeight: '100vh' }}>
            <Sider collapsible collapsed={collapsed} onCollapse={(value) => setCollapsed(value)} theme="dark">
                <div style={{ height: 32, margin: 16, background: 'rgba(255, 255, 255, 0.2)', borderRadius: 6 }} />
                <Menu theme="dark" defaultSelectedKeys={['users']} mode="inline" onSelect={(i) => setCurrentMenu(i.key)}>
                    <Menu.Item key="users" icon={<UserOutlined />}>Users</Menu.Item>
                    <Menu.Item key="papers" icon={<FileTextOutlined />}>Papers</Menu.Item>
                    <Menu.Item key="flags" icon={<ExclamationCircleOutlined />}>Flags/Reports</Menu.Item>
                </Menu>
            </Sider>
            <Layout>
                <Header style={{ padding: '0 24px', background: '#fff', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Title level={4} style={{ margin: 0 }}>Management Console</Title>
                    <Button type="text" icon={<LogoutOutlined />} onClick={handleLogout}>Logout</Button>
                </Header>
                <Content style={{ margin: '16px' }}>
                    <div style={{ padding: 24, minHeight: 360, background: '#fff', borderRadius: 8 }}>
                        {renderContent()}
                    </div>
                </Content>
            </Layout>
        </Layout>
    );
}

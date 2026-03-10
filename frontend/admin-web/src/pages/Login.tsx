import { useState } from 'react';
import { Form, Input, Button, Card, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import api from '../utils/api';

const { Title } = Typography;

export default function Login() {
    const [loading, setLoading] = useState(false);
    const navigate = useNavigate();

    const onFinish = async (values: any) => {
        setLoading(true);
        try {
            const res: any = await api.post('/login', values);
            localStorage.setItem('adminToken', res.token);
            message.success('Login successful');
            navigate('/');
        } catch {
            // Error handled by interceptor
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#f0f2f5' }}>
            <Card style={{ width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
                <div style={{ textAlign: 'center', marginBottom: 24 }}>
                    <Title level={3}>S.H.I.T. Admin Station</Title>
                </div>
                <Form name="login" onFinish={onFinish} size="large">
                    <Form.Item name="username" rules={[{ required: true, message: 'Please input username!' }]}>
                        <Input prefix={<UserOutlined />} placeholder="Admin Username" />
                    </Form.Item>
                    <Form.Item name="password" rules={[{ required: true, message: 'Please input password!' }]}>
                        <Input.Password prefix={<LockOutlined />} placeholder="Password" />
                    </Form.Item>
                    <Form.Item>
                        <Button type="primary" htmlType="submit" loading={loading} block>
                            Sign in
                        </Button>
                    </Form.Item>
                </Form>
            </Card>
        </div>
    );
}

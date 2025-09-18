import React from 'react';
import { Link, RouteComponentProps } from 'react-router-dom';
import './Register.less';
import TankComponent from '../../common/component/TankComponent';
import { Button, Col, Form, Input, Row, Select } from 'antd';
import User from '../../common/model/user/User';
import Moon from '../../common/model/global/Moon';
import { LockOutlined, UserOutlined, IdcardOutlined, PhoneOutlined, BankOutlined } from '@ant-design/icons';
import MessageBoxUtil from '../../common/util/MessageBoxUtil';
import Lang from '../../common/model/global/Lang';
import College from '../../common/model/college/College';

interface IProps extends RouteComponentProps {}

interface IState {
  colleges: any[];
}

export default class Register extends TankComponent<IProps, IState> {
  user: User = Moon.getSingleton().user;
  collegeModel: College = new College(this);

  constructor(props: IProps) {
    super(props);

    this.state = {
      colleges: [],
    };
  }

  componentDidMount() {
    // 加载学院列表
    this.collegeModel.httpList(
      () => {
        this.setState({
          colleges: this.collegeModel.colleges,
        });
      }
    );
  }

  onFinish(values: any) {
    let that = this;

    let user = that.user;

    // 构建用户档案信息
    const profileInfo = {
      realName: values['realName'],
      studentId: values['studentId'],
      college: values['college'],
      phoneNumber: values['phoneNumber'],
      userType: values['userType'],
    };

    user.httpRegister(values['username'], values['password'], profileInfo, function () {
      MessageBoxUtil.success(Lang.t('user.loginSuccess'));

      that.props.history.push('/');
    });
  }

  onFinishFailed(errorInfo: any) {}

  render() {
    let that = this;

    return (
      <div className="user-register">
        <Row>
          <Col xs={{ span: 24 }} md={{ span: 8, offset: 6 }}>
            <div className="welcome">{Lang.t('user.welcomeRegister')}</div>

            <Form
              name="basic"
              initialValues={{ remember: true }}
              onFinish={this.onFinish.bind(this)}
              onFinishFailed={this.onFinishFailed.bind(this)}
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: Lang.t('user.enterUsername') },
                ]}
              >
                <Input
                  size="large"
                  placeholder={Lang.t('user.enterUsername')}
                  prefix={<UserOutlined />}
                />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[
                  { required: true, message: Lang.t('user.enterPassword') },
                ]}
              >
                <Input.Password
                  size="large"
                  placeholder={Lang.t('user.enterPassword')}
                  prefix={<LockOutlined />}
                />
              </Form.Item>

              <Form.Item
                name="rePassword"
                dependencies={['password']}
                hasFeedback
                rules={[
                  {
                    required: true,
                    message: Lang.t('user.enterPassword'),
                  },
                  ({ getFieldValue }) => ({
                    validator(rule, value) {
                      if (!value || getFieldValue('password') === value) {
                        return Promise.resolve();
                      }
                      return Promise.reject(Lang.t('user.passwordNotSame'));
                    },
                  }),
                ]}
              >
                <Input.Password
                  size="large"
                  placeholder={Lang.t('user.confirmPassword')}
                  prefix={<LockOutlined />}
                />
              </Form.Item>

              {/* 个人信息字段 */}
              <Form.Item
                name="realName"
                rules={[
                  { required: true, message: '请输入真实姓名' },
                ]}
              >
                <Input
                  size="large"
                  placeholder="请输入真实姓名"
                  prefix={<UserOutlined />}
                />
              </Form.Item>

              <Form.Item
                name="studentId"
                rules={[
                  { required: true, message: '请输入学号/工号' },
                ]}
              >
                <Input
                  size="large"
                  placeholder="请输入学号/工号"
                  prefix={<IdcardOutlined />}
                />
              </Form.Item>

              <Form.Item
                name="college"
                rules={[
                  { required: true, message: '请选择学院' },
                ]}
              >
                <Select
                  size="large"
                  placeholder="请选择学院"
                >
                  {this.state.colleges.map((college: any) => (
                    <Select.Option key={college.id} value={college.name}>
                      {college.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                name="phoneNumber"
                rules={[
                  { required: true, message: '请输入手机号' },
                  { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
                ]}
              >
                <Input
                  size="large"
                  placeholder="请输入手机号"
                  prefix={<PhoneOutlined />}
                />
              </Form.Item>

              <Form.Item
                name="userType"
                rules={[
                  { required: true, message: '请选择身份' },
                ]}
              >
                <Select
                  size="large"
                  placeholder="请选择身份"
                >
                  <Select.Option value="student">学生</Select.Option>
                  <Select.Option value="teacher">教师</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item>
                <Button
                  type="primary"
                  icon={<UserOutlined />}
                  block={true}
                  htmlType="submit"
                >
                  {Lang.t('user.register')}
                </Button>
              </Form.Item>

              <div className="text-right">
                <Link to={'/user/login'}>
                  <span className="link">{Lang.t('user.goToLogin')}</span>
                </Link>
              </div>
            </Form>
          </Col>
        </Row>
      </div>
    );
  }
}

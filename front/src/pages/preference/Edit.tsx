import React from 'react';
import { RouteComponentProps } from 'react-router-dom';
import './Edit.less';
import TankComponent from '../../common/component/TankComponent';
import User from '../../common/model/user/User';
import Moon from '../../common/model/global/Moon';
import TankTitle from '../widget/TankTitle';
import { Button, Form, Input, InputNumber, Switch, Select, Space, Tabs, Table, Modal, message } from 'antd';
import { SaveOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import TankContentCard from '../widget/TankContentCard';
import Preference from '../../common/model/preference/Preference';
import { FormInstance } from 'antd/lib/form';
import MessageBoxUtil from '../../common/util/MessageBoxUtil';
import Sun from '../../common/model/global/Sun';
import Lang from '../../common/model/global/Lang';
import MatterImage from '../matter/widget/MatterImage';
import InputSize from '../widget/form/InputSize';
import College from '../../common/model/college/College';
import Track from '../../common/model/track/Track';

interface IProps extends RouteComponentProps {}

interface IState {
  collegeModalVisible: boolean;
  trackModalVisible: boolean;
  bulkCollegeModalVisible: boolean;
  newCollegeName: string;
  newTrackName: string;
  newTrackUserType: string;
  bulkCollegeNames: string;
}

export default class Edit extends TankComponent<IProps, IState> {
  formRef = React.createRef<FormInstance>();

  user: User = Moon.getSingleton().user;
  preference: Preference = Moon.getSingleton().preference;
  college: College = new College();
  track: Track = new Track();

  constructor(props: IProps) {
    super(props);

    this.state = {
      collegeModalVisible: false,
      trackModalVisible: false,
      bulkCollegeModalVisible: false,
      newCollegeName: '',
      newTrackName: '',
      newTrackUserType: 'BOTH',
      bulkCollegeNames: '',
    };

    this.preference.detailLoading = true;
  }

  componentDidMount() {
    let that = this;

    that.refreshPreference();
    that.refreshColleges();
    that.refreshTracks();
  }

  refreshPreference() {
    let that = this;

    that.preference.httpFetch(function () {
      that.preference.detailLoading = false;
      that.updateUI();
    });
  }

  refreshColleges() {
    let that = this;
    that.college.httpList(function () {
      that.updateUI();
    });
  }

  refreshTracks() {
    let that = this;
    that.track.httpList(function () {
      that.updateUI();
    });
  }

  onFinish(values: any) {
    let that = this;

    let user = that.user;

    that.preference.assign(values);

    that.preference.httpSave(function () {
      MessageBoxUtil.success(Lang.t('operationSuccess'));

      Sun.updateFrame();

      Sun.navigateTo('/preference/index');
    });
  }

  showCollegeModal = () => {
    this.setState({ collegeModalVisible: true, newCollegeName: '' });
  };

  hideCollegeModal = () => {
    this.setState({ collegeModalVisible: false });
  };

  showBulkCollegeModal = () => {
    this.setState({ bulkCollegeModalVisible: true, bulkCollegeNames: '' });
  };

  hideBulkCollegeModal = () => {
    this.setState({ bulkCollegeModalVisible: false });
  };

  showTrackModal = () => {
    this.setState({ 
      trackModalVisible: true, 
      newTrackName: '', 
      newTrackUserType: 'BOTH',
    });
  };

  hideTrackModal = () => {
    this.setState({ trackModalVisible: false });
  };

  handleCreateCollege = () => {
    const { newCollegeName } = this.state;
    if (!newCollegeName.trim()) {
      message.error('请输入学院名称');
      return;
    }

    this.college.httpCreate(newCollegeName, () => {
      message.success('学院创建成功');
      this.hideCollegeModal();
      this.refreshColleges();
    });
  };

  handleBulkCreateColleges = () => {
    const { bulkCollegeNames } = this.state;
    if (!bulkCollegeNames.trim()) {
      message.error('请输入学院名称');
      return;
    }

    this.college.httpBulkCreate(bulkCollegeNames, () => {
      message.success('学院批量创建成功');
      this.hideBulkCollegeModal();
      this.refreshColleges();
    });
  };

  handleCreateTrack = () => {
    const { newTrackName, newTrackUserType } = this.state;
    if (!newTrackName.trim()) {
      message.error('请输入赛道名称');
      return;
    }

    this.track.httpCreate(newTrackName, newTrackUserType, '', () => {
      message.success('赛道创建成功');
      this.hideTrackModal();
      this.refreshTracks();
    });
  };

  handleDeleteCollege = (collegeId: number) => {
    this.college.httpDelete(collegeId, () => {
      message.success('学院删除成功');
      this.refreshColleges();
    });
  };

  handleDeleteTrack = (trackId: number) => {
    this.track.httpDelete(trackId, () => {
      message.success('赛道删除成功');
      this.refreshTracks();
    });
  };

  onFinishFailed(errorInfo: any) {}

  render() {
    let that = this;

    let preference = that.preference;
    let colleges = that.college.colleges;
    let tracks = that.track.tracks;
    const { 
      collegeModalVisible, 
      trackModalVisible, 
      bulkCollegeModalVisible,
      newCollegeName, 
      newTrackName, 
      newTrackUserType, 
      bulkCollegeNames
    } = this.state;

    const layout = {
      labelCol: { span: 6 },
      wrapperCol: { span: 18 },
    };

    let initialValues: any = preference.getForm();

    const collegeColumns = [
      {
        title: '学院名称',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '创建时间',
        dataIndex: 'createTime',
        key: 'createTime',
        render: (text: string) => new Date(text).toLocaleString(),
      },
      {
        title: '操作',
        key: 'action',
        render: (text: any, record: any) => (
          <Button 
            type="link" 
            danger 
            icon={<DeleteOutlined />}
            onClick={() => this.handleDeleteCollege(record.id)}
          >
            删除
          </Button>
        ),
      },
    ];

    const trackColumns = [
      {
        title: '赛道名称',
        dataIndex: 'name',
        key: 'name',
      },
      {
        title: '适用用户',
        dataIndex: 'targetUserType',
        key: 'targetUserType',
        render: (text: string) => (
          text === 'TEACHER' ? '教师' : text === 'STUDENT' ? '学生' : '通用'
        ),
      },
      {
        title: '创建时间',
        dataIndex: 'createTime',
        key: 'createTime',
        render: (text: string) => new Date(text).toLocaleString(),
      },
      {
        title: '操作',
        key: 'action',
        render: (text: any, record: any) => (
          <Button 
            type="link" 
            danger 
            icon={<DeleteOutlined />}
            onClick={() => this.handleDeleteTrack(record.id)}
          >
            删除
          </Button>
        ),
      },
    ];

    return (
      <div className="page-preference-edit">
        <TankTitle name={Lang.t('preference.editPreference')} />

        <TankContentCard loading={preference.detailLoading}>
          <Tabs defaultActiveKey="basic">
            <Tabs.TabPane tab="基本设置" key="basic">
              <Form
                {...layout}
                name="basic"
                ref={this.formRef}
                initialValues={initialValues}
                onFinish={this.onFinish.bind(this)}
                onFinishFailed={this.onFinishFailed.bind(this)}
                onValuesChange={() => that.updateUI()}
              >
            <Form.Item
              label={Lang.t('preference.websiteName')}
              name="name"
              rules={[
                {
                  required: true,
                  message: Lang.t('preference.enterWebsiteName'),
                },
              ]}
            >
              <Input />
            </Form.Item>

            <Form.Item label="logo" name="logoUrl">
              <MatterImage uploadHint={Lang.t('preference.logoSquare')} />
            </Form.Item>

            <Form.Item label="favicon" name="faviconUrl">
              <MatterImage
                filter=".ico"
                previewWidth={60}
                uploadHint={Lang.t('preference.onlyAllowIco')}
              />
            </Form.Item>

            <Form.Item label={Lang.t('preference.copyright')} name="copyright">
              <Input />
            </Form.Item>

            <Form.Item label={Lang.t('preference.extraInfo')} name="record">
              <Input />
            </Form.Item>

            <Form.Item
              label={Lang.t('preference.zipMaxNumLimit')}
              name="downloadDirMaxNum"
              rules={[
                {
                  required: true,
                  message: Lang.t('preference.enterZipMaxNumLimit'),
                },
              ]}
            >
              <InputNumber min={-1} max={1000} className="w150" />
            </Form.Item>

            <Form.Item
              label={Lang.t('preference.zipMaxSizeLimit')}
              required={true}
              name="downloadDirMaxSize"
              rules={[
                {
                  required: true,
                  message: Lang.t('preference.enterZipMaxSizeLimit'),
                },
              ]}
            >
              <InputSize className="w200" />
            </Form.Item>

            <Form.Item
              label={Lang.t('preference.userDefaultSizeLimit')}
              required={true}
              name="defaultTotalSizeLimit"
              rules={[
                {
                  required: true,
                  message: Lang.t('preference.enterUserDefaultSizeLimit'),
                },
              ]}
            >
              <InputSize className="w200" />
            </Form.Item>

            <Form.Item
              label={Lang.t('preference.matterBinDefaultSaveDay')}
              required={true}
            >
              <Form.Item
                name="deletedKeepDays"
                rules={[
                  {
                    required: true,
                    message: Lang.t('preference.enterMatterBinDefaultSaveDay'),
                  },
                ]}
                noStyle
              >
                <InputNumber min={0} className="w150" />
              </Form.Item>
              <span className="pl10">
                {Lang.t('preference.matterBinDefaultTip')}
              </span>
            </Form.Item>

            <Form.Item
              label={Lang.t('preference.allowRegister')}
              name="allowRegister"
              required={true}
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>


            <div className="text-right">
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />}>
                {Lang.t('save')}
              </Button>
            </div>
          </Form>
            </Tabs.TabPane>

            <Tabs.TabPane tab="学院管理" key="college">
              <div style={{ marginBottom: 16 }}>
                <Space>
                  <Button 
                    type="primary" 
                    icon={<PlusOutlined />}
                    onClick={this.showCollegeModal}
                  >
                    添加学院
                  </Button>
                  <Button 
                    type="default" 
                    icon={<PlusOutlined />}
                    onClick={this.showBulkCollegeModal}
                  >
                    批量添加学院
                  </Button>
                </Space>
              </div>
              <Table 
                columns={collegeColumns} 
                dataSource={colleges} 
                rowKey="id"
                pagination={false}
              />
            </Tabs.TabPane>

            <Tabs.TabPane tab="赛道管理" key="track">
              <div style={{ marginBottom: 16 }}>
                <Button 
                  type="primary" 
                  icon={<PlusOutlined />}
                  onClick={this.showTrackModal}
                >
                  添加赛道
                </Button>
              </div>
              <Table 
                columns={trackColumns} 
                dataSource={tracks} 
                rowKey="id"
                pagination={false}
              />
            </Tabs.TabPane>
          </Tabs>
        </TankContentCard>

        <Modal
          title="添加学院"
          visible={collegeModalVisible}
          onOk={this.handleCreateCollege}
          onCancel={this.hideCollegeModal}
        >
          <Input
            placeholder="请输入学院名称"
            value={newCollegeName}
            onChange={(e) => this.setState({ newCollegeName: e.target.value })}
          />
        </Modal>

        <Modal
          title="添加赛道"
          visible={trackModalVisible}
          onOk={this.handleCreateTrack}
          onCancel={this.hideTrackModal}
          width={400}
        >
          <Form layout="vertical">
            <Form.Item label="赛道名称" required>
              <Input
                placeholder="请输入赛道名称"
                value={newTrackName}
                onChange={(e) => this.setState({ newTrackName: e.target.value })}
              />
            </Form.Item>
            <Form.Item label="适用用户类型" required>
              <Select
                value={newTrackUserType}
                onChange={(value) => this.setState({ newTrackUserType: value })}
              >
                <Select.Option value="BOTH">通用</Select.Option>
                <Select.Option value="TEACHER">教师</Select.Option>
                <Select.Option value="STUDENT">学生</Select.Option>
              </Select>
            </Form.Item>
          </Form>
        </Modal>

        <Modal
          title="批量添加学院"
          visible={bulkCollegeModalVisible}
          onOk={this.handleBulkCreateColleges}
          onCancel={this.hideBulkCollegeModal}
          width={600}
        >
          <Form layout="vertical">
            <Form.Item label="学院名称列表" required>
              <Input.TextArea
                placeholder="请输入学院名称，每行一个"
                value={bulkCollegeNames}
                onChange={(e) => this.setState({ bulkCollegeNames: e.target.value })}
                rows={6}
              />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    );
  }
}

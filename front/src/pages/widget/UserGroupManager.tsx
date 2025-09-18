import React, { useState, useEffect } from 'react';
import {
  Modal,
  Button,
  Form,
  Input,
  Checkbox,
  Table,
  Tag,
  message,
  Tooltip,
  Space,
} from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons';
import HttpUtil from '../../common/util/HttpUtil';
import { ColumnProps } from 'antd/es/table';

// 定义接口
interface Label {
  id: number;
  name: string;
  type: 'bool' | 'number';
}

interface UserGroup {
  id: number;
  name: string;
  display: string;
  editable_labels: string;
  editable: boolean;
}

interface FormValue {
  name: string;
  display: string[];
  editable_labels: string[];
  editable: boolean;
}

const UserGroupManager: React.FC = () => {
  const [isModalVisible, setIsModalVisible] = useState<boolean>(false);
  const [userGroups, setUserGroups] = useState<UserGroup[]>([]);
  const [labels, setLabels] = useState<Label[]>([]);
  const [idCnt, setIdCnt] = useState<number>(1);
  const [form] = Form.useForm<FormValue>();

  // 获取标签数据
  useEffect(() => {
    HttpUtil.httpGet(
      '/api/user/label',
      {},
      (res: { data: { data: Label[] } }) => {
        console.log(res);
        let labels = [...res.data.data];
        let curId = idCnt;
        labels.forEach((l) => {
          l.id = curId;
          curId += 1;
        });
        setIdCnt(curId + 1);
        setLabels(labels);
      }
    );
  }, []);

  useEffect(() => {
    HttpUtil.httpGet(
      '/api/user/group',
      {},
      (res: { data: { data: UserGroup[] } }) => {
        console.log(res);
        let groups = [...res.data.data];
        let curId = idCnt;
        groups.forEach((l) => {
          l.id = curId;
          curId += 1;
        });
        setIdCnt(curId + 1);
        setUserGroups(groups);
      }
    );
  }, []);

  const showModal = (): void => {
    setIsModalVisible(true);
  };

  const handleCancel = (): void => {
    setIsModalVisible(false);
    form.resetFields();
  };

  const onFinish = (values: FormValue): void => {
    if (userGroups.some((group) => group.name === values.name)) {
      message.error('用户组名称必须唯一');
      return;
    }

    const newUserGroup: UserGroup = {
      id: Date.now(),
      name: values.name,
      display: JSON.stringify(values.display),
      editable_labels: JSON.stringify(values.editable_labels),
      editable: !values.display || values.display.length == 0,
    };

    const newGroupJSON = JSON.stringify(newUserGroup);
    const jsonForm = new FormData();
    jsonForm.set('data', newGroupJSON);
    HttpUtil.httpPost(
      '/api/user/group/create',
      jsonForm,
      () => message.success(`用户组 "${values.name}" 添加成功`),
      (err: any) => message.error(`添加失败：${err}`)
    );
    setUserGroups([...userGroups, newUserGroup]);
    handleCancel();
  };

  const deleteUserGroup = (id: number): void => {
    const groupToDelete = userGroups.find((group) => group.id === id);
    if (groupToDelete) {
      const deleteGroupJSON = JSON.stringify(groupToDelete);
      const jsonForm = new FormData();
      jsonForm.set('data', deleteGroupJSON);
      HttpUtil.httpPost(
        '/api/user/group/delete',
        jsonForm,
        () => message.success(`删除成功`),
        (err: any) => message.error(`删除失败：${err}`)
      );
      setUserGroups(userGroups.filter((group) => group.id !== id));
    }
  };

  // 获取标签名称
  const getLabelName = (id: number): string => {
    const label = labels.find((l) => l.id === id);
    return label ? label.name : '无名标签';
  };

  // 表格列定义
  const columns: ColumnProps<NameUserGroup>[] = [
    {
      title: '用户组名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => <strong>{text}</strong>,
    },
    {
      title: '显示条件',
      dataIndex: 'displayConditions',
      key: 'displayConditions',
      render: (conditions: number[]) => (
        <div>
          {conditions.length > 0 ? (
            conditions.map((name) => (
              <Tag key={name} color="blue" style={{ margin: '2px' }}>
                {name}
              </Tag>
            ))
          ) : (
            <span style={{ color: '#999' }}>无</span>
          )}
        </div>
      ),
    },
    {
      title: '可编辑标签',
      dataIndex: 'editableLabels',
      key: 'editableLabels',
      render: (editableLabels: string[]) => (
        <div>
          {editableLabels.length > 0 ? (
            editableLabels.map((name) => (
              <Tag key={name} color="green" style={{ margin: '2px' }}>
                {name}
              </Tag>
            ))
          ) : (
            <span style={{ color: '#999' }}>无</span>
          )}
        </div>
      ),
    },
    {
      title: '可编辑内容',
      dataIndex: 'canEditContent',
      key: 'canEditContent',
      render: (canEdit: boolean) => (
        <Tag color={canEdit ? 'green' : 'red'}>{canEdit ? '是' : '否'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (text: any, record: NameUserGroup) => (
        <Space>
          <Tooltip title="删除">
            <Button
              type="text"
              icon={<DeleteOutlined />}
              danger
              onClick={() => deleteUserGroup(record.id)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  interface NameUserGroup {
    id: number;
    name: string;
    displayConditions: string[];
    editableLabels: string[];
    canEditContent: boolean;
  }

  const toLabelId = (group: UserGroup[]): NameUserGroup[] => {
    return group.map((g) => {
      const displayConditionsNames = (
        g.display ? JSON.parse(g.display) : []
      ) as string[];
      const editableLabelsNames = (
        g.editable_labels ? JSON.parse(g.editable_labels) : []
      ) as string[];
      return {
        id: g.id,
        name: g.name,
        displayConditions: displayConditionsNames,
        editableLabels: editableLabelsNames,
        canEditContent: g.editable,
      };
    });
  };

  return (
    <div style={{ paddingTop: '24px' }}>
      <div style={{ marginBottom: '16px' }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={showModal}>
          新建用户组
        </Button>
      </div>

      {/* 用户组表格 */}
      <Table
        rowKey="id"
        columns={columns}
        dataSource={toLabelId(userGroups)}
        locale={{
          emptyText: (
            <div
              style={{ textAlign: 'center', padding: '40px 0', color: '#999' }}
            >
              <p>暂无用户组，请创建新用户组</p>
            </div>
          ),
        }}
        pagination={userGroups.length > 10 ? { pageSize: 10 } : false}
      />

      {/* 新建用户组模态框 */}
      <Modal
        title="新建用户组"
        visible={isModalVisible}
        onCancel={handleCancel}
        footer={null}
        width={600}
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
        >
          <Form.Item
            label="用户组名称"
            name="name"
            rules={[
              { required: true, message: '请输入用户组名称' },
              { max: 20, message: '名称不能超过20个字符' },
            ]}
          >
            <Input placeholder="输入唯一用户组名称" />
          </Form.Item>

          <Form.Item label="显示条件（选择bool标签）" name="display">
            <Checkbox.Group>
              <Space direction="vertical">
                {labels
                  .filter((label) => label.type === 'bool')
                  .map((label) => (
                    <Checkbox key={label.id} value={label.name}>
                      {label.name}
                    </Checkbox>
                  ))}
              </Space>
            </Checkbox.Group>
          </Form.Item>

          <Form.Item label="可编辑标签" name="editable_labels">
            <Checkbox.Group>
              <Space direction="vertical">
                {labels.map((label) => (
                  <Checkbox key={label.id} value={label.name}>
                    {label.name} ({label.type})
                  </Checkbox>
                ))}
              </Space>
            </Checkbox.Group>
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                创建
              </Button>
              <Button onClick={handleCancel}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default UserGroupManager;

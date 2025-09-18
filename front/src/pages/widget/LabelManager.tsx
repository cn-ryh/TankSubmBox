import React, { useEffect, useState } from 'react';
import {
  Modal,
  Button,
  Form,
  Input,
  Select,
  Tag,
  Space,
  Card,
  message,
} from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import './LabelManager.less';
import HttpUtil from '../../common/util/HttpUtil';

const { Option } = Select;

// 定义标签接口
interface Label {
  id: number;
  name: string;
  type: 'bool' | 'numb';
}

// 定义表单值接口
interface FormValues {
  name: string;
  type: 'bool' | 'numb';
}

const LabelManager: React.FC = () => {
  const [isModalVisible, setIsModalVisible] = useState<boolean>(false);
  const [labels, setLabels] = useState<Label[]>([]);
  const [form] = Form.useForm<FormValues>();
  const [idCnt, setIdCnt] = useState<number>(0);

  useEffect(() => {
    HttpUtil.httpGet(
      '/api/user/label',
      {},
      (res: { data: { data: Label[] } }) => {
        console.log(res);
        let labels = [...res.data.data];
        let curId = idCnt;
        labels.forEach((l) => {
          l.id = curId
          curId += 1;
        });
        setIdCnt(curId + 1);
        setLabels(labels);
      }
    );
  }, []);

  const showModal = (): void => {
    setIsModalVisible(true);
  };

  const handleCancel = (): void => {
    setIsModalVisible(false);
  };

  const onFinish = (values: FormValues): void => {
    console.log(labels)
    const newLabel: Label = {
      id: idCnt,
      name: values.name,
      type: values.type,
    };
    setIdCnt(idCnt + 1);
    if (labels.find((l) => l.name == newLabel.name)) {
      message.error(`添加失败：标签名字不能重复`);
      return;
    }
    const formData = new FormData();
    formData.set('name', values.name);
    formData.set('type', values.type);
    console.log(values);
    HttpUtil.httpPost(
      '/api/user/label/create',
      formData,
      () => {
        message.success('标签添加成功');
      },
      (err: any) => {
        message.error(`标签添加失败${err}`);
      }
    );
    setLabels([...labels, newLabel]);

    console.log('添加标签:', newLabel);
    form.resetFields();
  };

  const deleteLabel = (id: number): void => {
    const labelToDelete = labels.find((label) => label.id === id);
    if (labelToDelete) {
      const deleteForm = new FormData();
      deleteForm.set('name', labelToDelete.name);
      deleteForm.set('type', labelToDelete.type);
      setLabels(labels.filter((label) => label.id !== id));
      HttpUtil.httpPost(
        '/api/user/label/delete',
        deleteForm,
        () => {
          message.success('标签删除成功');
        },
        (err: any) => {
          message.error(`标签删除失败${err}`);
        }
      );
    }
  };

  return (
    <div className="label-manager-container">
      <Button
        type="primary"
        icon={<PlusOutlined />}
        onClick={showModal}
        className="ml10"
      >
        管理标签
      </Button>

      <Modal
        title="标签管理"
        visible={isModalVisible}
        onCancel={handleCancel}
        footer={[
          <Button key="close" onClick={handleCancel}>
            关闭
          </Button>,
        ]}
        width={600}
      >
        <div className="modal-content">
          <Card title="添加新标签" size="small" className="add-label-card">
            <Form<FormValues>
              form={form}
              layout="vertical"
              onFinish={onFinish}
              autoComplete="off"
            >
              <Space align="start" className="form-row">
                <Form.Item
                  label="标签名称"
                  name="name"
                  rules={[{ required: true, message: '请输入标签名称' }]}
                >
                  <Input placeholder="输入标签名称" />
                </Form.Item>

                <Form.Item
                  label="标签类型"
                  name="type"
                  rules={[{ required: true, message: '请选择标签类型' }]}
                >
                  <Select style={{ width: 120 }} placeholder="选择类型">
                    <Option value="bool">有无</Option>
                    <Option value="numb">打分</Option>
                  </Select>
                </Form.Item>

                <Form.Item label=" " style={{ marginBottom: 0 }}>
                  <Button
                    type="primary"
                    htmlType="submit"
                    icon={<PlusOutlined />}
                  >
                    添加
                  </Button>
                </Form.Item>
              </Space>
            </Form>
          </Card>

          <Card title="已添加的标签" className="labels-list-card">
            {labels.length === 0 ? (
              <div className="empty-state">
                <p>暂无标签，请添加新标签</p>
              </div>
            ) : (
              <div className="labels-container">
                {labels.map((label) => (
                  <Tag
                    key={label.id}
                    color={label.type === 'bool' ? 'blue' : 'green'}
                    closable
                    onClose={() => deleteLabel(label.id)}
                    className="label-tag"
                  >
                    {label.name} ({label.type})
                  </Tag>
                ))}
              </div>
            )}
          </Card>
        </div>
      </Modal>
    </div>
  );
};

export default LabelManager;

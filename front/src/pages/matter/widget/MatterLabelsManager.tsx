import React, { useEffect, useState } from 'react';
import {
  Tooltip,
  Modal,
  Select,
  InputNumber,
  Tag,
  Button,
  Space,
  message,
} from 'antd';
import { LinkOutlined, CloseOutlined, TagOutlined } from '@ant-design/icons';
import { Label } from '../../../common/model/user/UserRole';
import SafeUtil from '../../../common/util/SafeUtil';
import HttpUtil from '../../../common/util/HttpUtil';

interface AddedLabel extends Label {
  value?: number;
}

interface LabelManagerProps {
  allLabels: Label[];
  uuid: string;
  userUuid: string;
}

const LabelManager: React.FC<LabelManagerProps> = ({
  allLabels,
  uuid,
  userUuid,
}) => {
  const [visible, setVisible] = useState<boolean>(false);
  const [selectedLabelId, setSelectedLabelId] = useState<number | null>(null);
  const [numbValue, setNumbValue] = useState<number>(1);
  const [existingLabels, setExistingLabels] = useState<AddedLabel[]>([]);

  useEffect(() => {
    let cnt = 0;
    const query = new FormData();
    query.set('uuid', uuid);
    console.log(query.get("uuid"))
    HttpUtil.httpGet(
      '/api/matter/label',
      {uuid},
      (res: { data: { data: AddedLabel[] } }) => {
        console.log(res);
        const labels = res.data.data;
        labels.forEach((l) => (l.id = ++cnt));
        setExistingLabels(labels);
      }
    );
  }, []);

  const selectedLabel = selectedLabelId != null
    ? allLabels.find((label) => label.id === selectedLabelId)
    : null;

  const showModal = (): void => {
    setVisible(true);
  };

  const hideModal = (): void => {
    setVisible(false);
    setSelectedLabelId(null);
    setNumbValue(1);
  };

  const handleAddLabel = (): void => {
    if (selectedLabelId == null || !selectedLabel) {
      message.warning('请先选择一个标签');
      return;
    }

    if (selectedLabel.type === 'numb' && (numbValue < 1 || numbValue > 100)) {
      message.error('请输入1-100之间的数字');
      return;
    }

    const newLabel: AddedLabel = {
      ...selectedLabel,
      ...(selectedLabel.type === 'numb' && { value: numbValue }),
    };

    const exists = existingLabels.some((label) => {
      return label.name == newLabel.name;
    });

    if (exists) {
      message.warning('该标签已存在');
      return;
    }

    const updatedLabels = [...existingLabels, newLabel];
    console.log('添加标签:', newLabel);
    console.log('当前所有标签:', updatedLabels);

    const update = new FormData();
    update.set('target', uuid);
    update.set('labelName', newLabel.name);
    update.set('value', String(newLabel.value ?? 0));
    update.set('userUuid', userUuid);
    HttpUtil.httpPost(
      '/api/matter/label/create',
      update,
      () => {
        setExistingLabels(updatedLabels);
      },
      (err: any) => message.error(err)
    );

    setSelectedLabelId(null);
    setNumbValue(1);
  };

  const handleDeleteLabel = (labelToDelete: AddedLabel): void => {
    const updatedLabels = existingLabels.filter((label) => {
      if (label.id !== labelToDelete.id) return true;

      if (label.type === 'numb' && labelToDelete.type === 'numb') {
        return label.value !== labelToDelete.value;
      }

      return false;
    });

    const update = new FormData();
    update.set('target', uuid);
    update.set('labelName', labelToDelete.name);
    HttpUtil.httpPost('/api/matter/label/delete', update, () => {
      setExistingLabels(updatedLabels);
    });
  };

  const handleLabelSelect = (value: number): void => {
    console.log(value);
    setSelectedLabelId(value);
  };

  const wrapStopPropagation = (handler: Function) => (e: React.MouseEvent) => {
    SafeUtil.stopPropagationWrap(e)(handler());
  };

  return (
    <>
      <Tooltip title="管理标签">
        <TagOutlined
          className="btn-action"
          onClick={(e) => SafeUtil.stopPropagationWrap(e)(showModal())}
          style={{ fontSize: '16px', color: '#1890ff', cursor: 'pointer' }}
        />
      </Tooltip>

      <div
        onClick={(e) => SafeUtil.stopPropagationWrap(e)}
        style={{ display: 'inline-block' }}
      >
        <Modal
          title="标签管理"
          visible={visible}
          onCancel={wrapStopPropagation(hideModal)}
          footer={[
            <Button key="cancel" onClick={wrapStopPropagation(hideModal)}>
              取消
            </Button>,
            <Button
              key="add"
              type="primary"
              onClick={wrapStopPropagation(handleAddLabel)}
              disabled={selectedLabelId == null}
            >
              添加
            </Button>,
          ]}
          width={500}
        >
          {/* 添加标签区域 */}
          <div
            style={{ marginBottom: 20 }}
            onClick={wrapStopPropagation(() => {})}
          >
            <Space>
              <Select
                placeholder="选择标签"
                style={{ width: 200 }}
                onChange={handleLabelSelect}
                value={selectedLabelId}
                allowClear
                onClick={wrapStopPropagation(() => {})}
              >
                {allLabels.map((label) => (
                  <Select.Option
                    key={label.id}
                    value={label.id}
                    onClick={wrapStopPropagation(() =>
                      setSelectedLabelId(label.id)
                    )}
                  >
                    {label.name}
                  </Select.Option>
                ))}
              </Select>

              {selectedLabel && selectedLabel.type === 'numb' && (
                <InputNumber
                  min={1}
                  max={100}
                  value={numbValue}
                  onChange={(value) => setNumbValue(value || 1)}
                  placeholder="输入1-100的值"
                  onClick={wrapStopPropagation(() => {})}
                />
              )}
            </Space>
          </div>

          {/* 分隔线 */}
          <hr
            style={{ margin: '20px 0' }}
            onClick={wrapStopPropagation(() => {})}
          />

          {/* 现有标签区域 */}
          <div onClick={wrapStopPropagation(() => {})}>
            <h4>现有标签</h4>
            <div style={{ minHeight: 60 }}>
              {existingLabels.length === 0 ? (
                <p style={{ color: '#999' }}>暂无标签，请添加</p>
              ) : (
                <Space wrap>
                  {existingLabels.map((label, index) => (
                    <Tag
                      key={`${label.id}-${label.value || 'bool'}-${index}`}
                      closable
                      closeIcon={<CloseOutlined />}
                      onClose={wrapStopPropagation(() =>
                        handleDeleteLabel(label)
                      )}
                      style={{
                        padding: '5px 10px',
                        borderRadius: 4,
                        display: 'flex',
                        alignItems: 'center',
                        gap: '4px',
                      }}
                      onClick={wrapStopPropagation(() => {})}
                    >
                      {label.name}
                      {label.type === 'numb' && `: ${label.value}`}
                    </Tag>
                  ))}
                </Space>
              )}
            </div>
          </div>
        </Modal>
      </div>
    </>
  );
};

export default LabelManager;

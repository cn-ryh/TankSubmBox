import React from 'react';
import TankComponent from '../../../common/component/TankComponent';
import Matter from '../../../common/model/matter/Matter';
import { Modal, InputNumber, Input, Form, Button } from 'antd';
import HttpUtil from '../../../common/util/HttpUtil';
import MessageBoxUtil from '../../../common/util/MessageBoxUtil';

interface IProps {
  matter: Matter;
  submissionId: number;
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

interface IState {
  loading: boolean;
  score: number | null;
  comment: string;
}

export default class RatingModal extends TankComponent<IProps, IState> {
  constructor(props: IProps) {
    super(props);
    this.state = {
      loading: false,
      score: null,
      comment: ''
    };
  }

  componentDidUpdate(prevProps: IProps) {
    if (prevProps.visible !== this.props.visible && this.props.visible) {
      // 重置表单状态
      this.setState({
        score: null,
        comment: ''
      });
    }
  }

  // 提交评分
  submitRating = () => {
    const { score, comment } = this.state;
    const { submissionId } = this.props;

    if (score === null || score < 0 || score > 100) {
      MessageBoxUtil.error('请输入0-100之间的评分');
      return;
    }

    this.setState({ loading: true });

    const data = new FormData();
    data.set('submissionId', submissionId.toString());
    data.set('score', score.toString());
    data.set('comment', comment);

    HttpUtil.httpPost(
      '/api/rating/submit',
      data,
      (response: any) => {
        this.setState({ loading: false });
        MessageBoxUtil.success('评分成功');
        this.props.onSuccess();
        this.props.onCancel();
      },
      (msg: string) => {
        this.setState({ loading: false });
        MessageBoxUtil.error(msg || '评分失败');
      }
    );
  };

  // 处理评分变化
  handleScoreChange = (value: number | null) => {
    this.setState({ score: value });
  };

  // 处理评语变化
  handleCommentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    this.setState({ comment: e.target.value });
  };

  render() {
    const { matter, visible, onCancel } = this.props;
    const { loading, score, comment } = this.state;

    return (
      <Modal
        title={`评分 - ${matter.name}`}
        visible={visible}
        onCancel={onCancel}
        footer={[
          <Button key="cancel" onClick={onCancel}>
            取消
          </Button>,
          <Button
            key="submit"
            type="primary"
            loading={loading}
            onClick={this.submitRating}
            disabled={score === null || score < 0 || score > 100}
          >
            提交评分
          </Button>
        ]}
      >
        <Form layout="vertical">
          <Form.Item label="评分（0-100分）" required>
            <InputNumber
              min={0}
              max={100}
              value={score === null ? undefined : score}
              onChange={this.handleScoreChange}
              style={{ width: '100%' }}
              placeholder="请输入0-100之间的分数"
            />
          </Form.Item>
          <Form.Item label="评语（可选）">
            <Input.TextArea
              value={comment}
              onChange={this.handleCommentChange}
              placeholder="请输入评语"
              rows={4}
            />
          </Form.Item>
        </Form>
      </Modal>
    );
  }
}
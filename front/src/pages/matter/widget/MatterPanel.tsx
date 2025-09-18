import React from 'react';
import Matter from '../../../common/model/matter/Matter';
import TankComponent from '../../../common/component/TankComponent';
import Director from './Director';
import './MatterPanel.less';
import StringUtil from '../../../common/util/StringUtil';
import DateUtil from '../../../common/util/DateUtil';
import AnimateUtil from '../../../common/util/AnimateUtil';
import MessageBoxUtil from '../../../common/util/MessageBoxUtil';
import Expanding from '../../widget/Expanding';
import {
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  EllipsisOutlined,
  InfoCircleOutlined,
  LinkOutlined,
  LockOutlined,
  UnlockOutlined,
  StarOutlined,
} from '@ant-design/icons';
import { Checkbox, Dropdown, Menu, Tooltip, Modal } from 'antd';
import { CheckboxChangeEvent } from 'antd/es/checkbox';
import SafeUtil from '../../../common/util/SafeUtil';
import ClipboardUtil from '../../../common/util/ClipboardUtil';
import Lang from '../../../common/model/global/Lang';
import MatterDeleteModal from './MatterDeleteModal';
import { SpaceMemberRole } from '../../../common/model/space/member/SpaceMemberRole';
import { Label } from '../../../common/model/user/UserRole';
import HttpUtil from '../../../common/util/HttpUtil';
import Moon from '../../../common/model/global/Moon';
import { UserRole } from '../../../common/model/user/UserRole';
import RatingModal from './RatingModal';

interface IProps {
  matter: Matter;
  mode: 'normal' | 'space'; // normal:正常模式，space:空间模式
  spaceMemberRole?: SpaceMemberRole; // 用户在当前空间下的角色,只有在space模式下才有值
  director?: Director;
  allLabels: Label[];
  userUuid: string,
  onCreateDirectoryCallback?: () => any;
  onDeleteSuccess?: () => any;
  onCheckMatter?: (matter?: Matter) => any;
  onPreviewImage?: (matter: Matter) => any;
  onGoToDirectory?: (id: string) => any;
  onGoDetail?: (matter: Matter) => any;
  onRatingSuccess?: () => any;
}

interface IState {
  ratingModalVisible: boolean;
  submissionId: number | null;
}

export default class MatterPanel extends TankComponent<IProps, IState> {
  // 正在重命名的临时字段
  renameMatterName: string = '';
  // 正在向服务器提交rename的请求
  renamingLoading: boolean = false;
  // 小屏幕下操作栏
  showMore: boolean = false;

  inputRef = React.createRef<HTMLInputElement>();

  constructor(props: IProps) {
    super(props);
    this.state = {
      ratingModalVisible: false,
      submissionId: null
    };
  }

  // 如果在空间下，有些操作权限只对管理员和读写成员开放
  checkHandlePermission() {
    if (this.props.mode !== 'space') return true;
    return [SpaceMemberRole.ADMIN, SpaceMemberRole.READ_WRITE].includes(
      this.props.spaceMemberRole!
    );
  }

  // 检查是否是学院管理员
  isCollegeAdmin() {
    const user = Moon.getSingleton().user;
    return user.role === UserRole.COLLEGE_ADMIN;
  }

  // 检查是否是评委
  isJudge() {
    const user = Moon.getSingleton().user;
    return user.role === UserRole.JUDGE;
  }

  // 检查当前作品是否已经被评分
  isMatterRated(): boolean {
    // 这个方法不再使用，评分状态将通过其他方式处理
    return false;
  }

  // 评分
  scoreMatter() {
    const { matter } = this.props;
    
    // 获取提交ID
    const data = new FormData();
    data.set('matterUuid', matter.uuid ?? '');
    
    HttpUtil.httpPost(
      '/api/submission/by-matter',
      data,
      (response: any) => {
        if (response && response.data && response.data.id) {
          this.setState({
            ratingModalVisible: true,
            submissionId: response.data.id
          });
        } else {
          MessageBoxUtil.error('未找到对应的作品提交信息');
        }
      },
      (msg: string) => {
        MessageBoxUtil.error(msg || '获取作品提交信息失败');
      }
    );
  }

  // 关闭评分模态框
  closeRatingModal = () => {
    this.setState({
      ratingModalVisible: false,
      submissionId: null
    });
  }

  // 评分成功回调
  handleRatingSuccess = () => {
    // 调用父组件的回调函数
    if (this.props.onRatingSuccess) {
      this.props.onRatingSuccess();
    }
  }

  prepareRename() {
    const { matter, director } = this.props;
    if (director!.isEditing()) {
      console.error('导演正忙着，不予执行');
      return;
    }

    //告诉导演，自己正在编辑
    director!.renameMode = true;
    matter.editMode = true;
    this.renameMatterName = matter.name!;
    this.updateUI();
    setTimeout(() => {
      if (!this.inputRef.current) return;
      //如果是文件夹，全选中
      let dotIndex = matter.name!.lastIndexOf('.');
      if (dotIndex === -1) {
        AnimateUtil.setInputSelection(
          this.inputRef.current,
          0,
          this.renameMatterName.length
        );
      } else {
        AnimateUtil.setInputSelection(this.inputRef.current, 0, dotIndex);
      }
    });
  }

  clipboard() {
    let textToCopy = this.props.matter.getDownloadUrl();
    ClipboardUtil.copy(textToCopy, () => {
      MessageBoxUtil.success(Lang.t('operationSuccess'));
    });
  }

  deleteMatter() {
    MatterDeleteModal.open(
      () => {
        this.props.matter.httpSoftDelete(() => {
          MessageBoxUtil.success(Lang.t('operationSuccess'));
          this.props.onDeleteSuccess!();
        });
      },
      () => {
        this.props.matter.httpDelete(() => {
          MessageBoxUtil.success(Lang.t('operationSuccess'));
          this.props.onDeleteSuccess!();
        });
      }
    );
  }

  // 推荐作品
  recommendMatter() {
    const { matter } = this.props;
    const data = new FormData();
    data.set("matterUuid", matter.uuid ?? "");
    Modal.confirm({
      title: '确认推荐',
      content: '确定要推荐这个作品吗？',
      onOk: () => {
        HttpUtil.httpPost(
          `/api/submission/recommend`,
          data,
          () => {
            MessageBoxUtil.success('推荐成功');
          },
          (msg: string) => {
            MessageBoxUtil.error(msg || '推荐失败');
          }
        );
      }
    });
  }

  changeMatterName(e: any) {
    this.renameMatterName = e.currentTarget.value;
    this.updateUI();
  }

  finishRename() {
    //有可能按enter的时候和blur同时了。
    if (this.renamingLoading) {
      return;
    }
    const { matter, director } = this.props;
    this.renamingLoading = true;

    matter.httpRename(
      this.renameMatterName,
      () => {
        this.renamingLoading = false;
        MessageBoxUtil.success(Lang.t('operationSuccess'));
        //告诉导演，自己编辑完毕
        director!.renameMode = false;
        matter.editMode = false;
      },
      (msg: string) => {
        this.renamingLoading = false;
        MessageBoxUtil.error(msg);
        //告诉导演，自己编辑完毕
        director!.renameMode = false;
        matter.editMode = false;
      },
      () => this.updateUI()
    );
  }

  finishCreateDirectory() {
    const { matter, director, onCreateDirectoryCallback } = this.props;
    matter.name = this.renameMatterName;
    matter.httpCreateDirectory(
      () => {
        director!.createMode = false;
        matter.editMode = false;
        matter.assign(new Matter());
      },
      (msg: string) => {
        director!.createMode = false;
        matter.editMode = false;
        MessageBoxUtil.error(msg);
      },
      () => onCreateDirectoryCallback!()
    );
  }

  blurTrigger() {
    const { matter, director } = this.props;
    if (matter.editMode) {
      if (director!.createMode) {
        this.finishCreateDirectory();
      } else if (director!.renameMode) {
        this.finishRename();
      }
    }
    }

  enterTrigger(e: any) {
    if (e.key.toLowerCase() === 'enter') {
      this.inputRef.current!.blur();
    }
  }

  changePrivacy(privacy: boolean) {
    this.props.matter.httpChangePrivacy(privacy, () => {
      this.updateUI();
    });
  }

  checkToggle(e: CheckboxChangeEvent) {
    this.props.matter.check = e.target.checked;
    this.props.onCheckMatter!(this.props.matter);
  }

  highLight() {
    this.inputRef.current!.select();
  }

  clickRow() {
    const { matter, director, onGoToDirectory, onPreviewImage } =
      this.props;
    if (director && director.isEditing()) {
      console.error('导演正忙着，不予执行');
      return;
    }

    if (matter.dir) {
      onGoToDirectory!(matter.uuid!);
    } else {
      //图片进行预览操作
      if (matter.isImage()) {
        onPreviewImage!(matter);
      } else {
        matter.preview();
      }
    }
  }

  toggleHandles() {
    this.showMore = !this.showMore;
    this.updateUI();
  }

  renderPcOperation() {
    const { matter } = this.props;

    return (
      <div className="right-part">
        <span className="matter-operation text-theme">
          {/* 保留下载按钮 */}
          <Tooltip title={Lang.t('matter.download')}>
            <DownloadOutlined
              className="btn-action"
              onClick={(e) =>
                SafeUtil.stopPropagationWrap(e)(matter.download())
              }
            />
          </Tooltip>

          {/* 恢复删除按钮 */}
          {this.checkHandlePermission() && (
            <Tooltip title={Lang.t('matter.delete')}>
              <DeleteOutlined
                className="btn-action text-danger"
                onClick={(e) =>
                  SafeUtil.stopPropagationWrap(e)(this.deleteMatter())
                }
              />
            </Tooltip>
          )}

          {/* 学院管理员可以看到推荐按钮 */}
          {this.isCollegeAdmin() && (
            <Tooltip title="推荐作品">
              <StarOutlined
                className="btn-action"
                onClick={(e) =>
                  SafeUtil.stopPropagationWrap(e)(this.recommendMatter())
                }
              />
            </Tooltip>
          )}

          {/* 评委可以看到评分按钮 */}
          {this.isJudge() && this.props.matter.dir && (
            <Tooltip title="评分">
              <StarOutlined
                className="btn-action"
                onClick={(e) =>
                  SafeUtil.stopPropagationWrap(e)(this.scoreMatter())
                }
              />
            </Tooltip>
          )}
        </span>
        <Tooltip title={Lang.t('matter.size')}>
          <span className="matter-size">
            {StringUtil.humanFileSize(matter.size)}
          </span>
        </Tooltip>
        <Tooltip title={Lang.t('matter.updateTime')}>
          <span className="matter-date mr10">
            {DateUtil.simpleDateHourMinute(matter.updateTime)}
          </span>
        </Tooltip>
      </div>
    );
  }

  getHandles() {
    const { matter } = this.props;

    // 普通模式 - 返回菜单项配置而不是React元素
    const handles = [
      {
        key: 'detail',
        label: Lang.t('matter.fileDetail'),
        icon: <InfoCircleOutlined className="btn-action mr5" />,
        onClick: () => this.props.onGoDetail?.(matter)
      },
      ...(this.checkHandlePermission()
        ? [
            {
              key: 'rename',
              label: Lang.t('matter.rename'),
              icon: <EditOutlined className="btn-action mr5" />,
              onClick: () => this.prepareRename()
            }
          ]
        : []),
      ...(matter.dir
        ? []
        : [
            {
              key: 'copyLink',
              label: Lang.t('matter.copyLink'),
              icon: <LinkOutlined className="btn-action mr5" />,
              onClick: () => this.clipboard()
            }
          ]),
      {
        key: 'download',
        label: Lang.t('matter.download'),
        icon: <DownloadOutlined className="btn-action mr5" />,
        onClick: () => matter.download()
      },
      ...(this.isJudge() && matter.dir
        ? [
            {
              key: 'score',
              label: '评分',
              icon: <StarOutlined className="btn-action mr5" />,
              onClick: () => this.scoreMatter()
            }
          ]
        : []),
      ...(this.checkHandlePermission()
        ? [
            {
              key: 'delete',
              label: Lang.t('matter.delete'),
              icon: <DeleteOutlined className="btn-action mr5" />,
              onClick: () => this.deleteMatter()
            }
          ]
        : []),
    ];
    
    if (this.checkHandlePermission() && !matter.dir) {
      handles.unshift(
        matter.privacy ? 
        {
          key: 'setPublic',
          label: Lang.t('matter.setPublic'),
          icon: <UnlockOutlined className="btn-action mr5" />,
          onClick: () => this.changePrivacy(false)
        } : 
        {
          key: 'setPrivate',
          label: Lang.t('matter.setPrivate'),
          icon: <LockOutlined className="btn-action mr5" />,
          onClick: () => this.changePrivacy(true)
        }
      );
    }

    return handles;
  }

  renderMobileOperation() {
    const { matter } = this.props;

    return (
      <div className="more-panel">
        <div className="cell-btn navy text">
          <span>{DateUtil.simpleDateHourMinute(matter.updateTime)}</span>
          <span className="matter-size">
            {StringUtil.humanFileSize(matter.size)}
          </span>
        </div>
        {this.getHandles().map((item) => (
          <div
            key={item.key}
            className="cell-btn navy"
            onClick={(e) => {
              SafeUtil.stopPropagationWrap(e)(item.onClick?.());
            }}
          >
            {item.icon}
            {item.label}
          </div>
        ))}
      </div>
    );
  }

  render() {
    const { matter } = this.props;
    const { ratingModalVisible, submissionId } = this.state;

    const menu = (
      <Menu>
        {this.getHandles().map((item) => (
          <Menu.Item 
            key={item.key} 
            icon={item.icon}
            onClick={item.onClick}
          >
            {item.label}
          </Menu.Item>
        ))}
      </Menu>
    );

    return (
      <>
        <Dropdown overlay={menu} trigger={['contextMenu']}>
          <div className="widget-matter-panel">
            <div
              onClick={(e) => SafeUtil.stopPropagationWrap(e)(this.clickRow())}
            >
              <div className="media clearfix">
                <div className="pull-left">
                  <div className="left-part">
                    <span
                      className="cell cell-hot"
                      onClick={(e) => SafeUtil.stopPropagationWrap(e)}
                    >
                      <Checkbox
                        checked={matter.check}
                        onChange={(e) => this.checkToggle(e)}
                      />
                    </span>
                    <span className="cell">
                      <img className="matter-icon" src={matter.getIcon()} />
                    </span>
                  </div>
                </div>

                {/*在大屏幕下的操作栏*/}
                <div className="pull-right visible-pc">
                  {matter.uuid && this.renderPcOperation()}
                </div>

                <div className="pull-right visible-mobile">
                  <span
                    className="more-btn"
                    onClick={(e) =>
                      SafeUtil.stopPropagationWrap(e)(this.toggleHandles())
                    }
                  >
                    <EllipsisOutlined className="btn-action navy f18" />
                  </span>
                </div>

                <div className="media-body">
                  <div className="middle-part">
                    {matter.editMode ? (
                      <span className="matter-name-edit">
                        <input
                          ref={this.inputRef}
                          className={matter.uuid!}
                          value={this.renameMatterName}
                          onChange={(e) => this.changeMatterName(e)}
                          placeholder={Lang.t('matter.enterName')}
                          onBlur={() => this.blurTrigger()}
                          onKeyUp={(e) => this.enterTrigger(e)}
                        />
                      </span>
                    ) : (
                      <span className="matter-name">
                        {matter.name}
                        {!matter.dir && !matter.privacy && (
                          <Tooltip
                            title={Lang.t('matter.publicFileEveryoneCanVisit')}
                          >
                            <UnlockOutlined className="icon" />
                          </Tooltip>
                        )}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </div>

            <Expanding>
              {this.showMore ? this.renderMobileOperation() : null}
            </Expanding>
          </div>
        </Dropdown>

        {/* 评分模态框 */}
        {ratingModalVisible && submissionId !== null && (
          <RatingModal
            matter={matter}
            submissionId={submissionId as number}
            visible={ratingModalVisible}
            onCancel={this.closeRatingModal}
            onSuccess={this.handleRatingSuccess}
          />
        )}
      </>
    );
  }
}
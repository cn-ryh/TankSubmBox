import React, { Children } from 'react';
import './List.less';
import TankComponent from '../../common/component/TankComponent';
import Pager from '../../common/model/base/Pager';
import Matter from '../../common/model/matter/Matter';
import Moon from '../../common/model/global/Moon';
import Director from './widget/Director';
import SortDirection from '../../common/model/base/SortDirection';
import MatterPanel from './widget/MatterPanel';
import UploadMatterPanel from './widget/UploadMatterPanel';
import {
  Button,
  Col,
  Dropdown,
  Empty,
  Input,
  Menu,
  Modal,
  Pagination,
  Row,
  Space as AntdSpace,
  Upload,
  List as AntdList,
} from 'antd';
import MessageBoxUtil from '../../common/util/MessageBoxUtil';
import {
  CloudDownloadOutlined,
  CloudUploadOutlined,
  DeleteOutlined,
  DownloadOutlined,
  DragOutlined,
  ExclamationCircleFilled,
  FolderOutlined,
  MenuOutlined,
  MinusSquareOutlined,
  PlusSquareOutlined,
  SearchOutlined,
  ShareAltOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import ImagePreviewer from '../widget/previewer/ImagePreviewer';
import Sun from '../../common/model/global/Sun';
import StringUtil from '../../common/util/StringUtil';
import MoveBatchModal from './widget/MoveBatchModal';
import ShareOperationModal from './widget/ShareOperationModal';
import MatterDeleteModal from './widget/MatterDeleteModal';
import Share from '../../common/model/share/Share';
import BreadcrumbModel from '../../common/model/base/option/BreadcrumbModel';
import BreadcrumbPanel from '../widget/BreadcrumbPanel';
import Lang from '../../common/model/global/Lang';
import FileUtil from '../../common/util/FileUtil';
import SafeUtil from '../../common/util/SafeUtil';
import MatterSortPanel from './widget/MatterSortPanel';
import { UploadRequestOption as RcCustomRequestOptions } from 'rc-upload/lib/interface';
import Capacity from '../layout/widget/Capacity';
import Space from '../../common/model/space/Space';
import MatterCrawlModal from './widget/MatterCrawlModal';
import SpaceMember from '../../common/model/space/member/SpaceMember';
import { SpaceMemberRole } from '../../common/model/space/member/SpaceMemberRole';
import { debounce } from '../../common/util/OptimizeUtil';
import HttpUtil from '../../common/util/HttpUtil';
import { Label } from '../../common/model/user/UserRole';
import Track from '../../common/model/track/Track';
import { UserRole } from '../../common/model/user/UserRole';

interface IProps {
  spaceUuid?: string;
}

interface IState {
  trackId?: number;
  workName?: string;
  hasSubmission: boolean;
  tracks: Track[];
}

export default class List extends TankComponent<IProps, IState> {
  isInSpace = !!this.props.spaceUuid; // 是否在空间中
  // 当前空间
  space = new Space();
  // 当前用户在当前空间下的成员信息
  spaceMember = new SpaceMember();
  // 当前目录
  currentDirectory = new Matter();
  //准备新建的文件
  newMatter = new Matter();

  //当前选中的文件
  selectedMatters: Matter[] = [];
  //获取分页的一个帮助器
  pager = new Pager<Matter>(this, Matter, Pager.MAX_PAGE_SIZE);
  user = Moon.getSingleton().user;
  preference = Moon.getSingleton().preference;
  //导演
  director = new Director();
  //当前面包屑模型数组。
  breadcrumbModels: BreadcrumbModel[] = [];
  //分享
  share = new Share();

  newMatterRef = React.createRef<MatterPanel>();

  // 最外层div
  wrapperRef = React.createRef<HTMLDivElement>();
  uploadDirectoryBtnRef = React.createRef<HTMLElement>();

  //用来判断是否展示遮罩层
  dragEnterCount = 0;

  //临时文件list，用于上传文件夹功能
  tempUploadList: File[] = [];

  //上传错误日志
  uploadErrorLogs: [string, string, string][] = [];

  // 全局的上传对象
  uploadMattersMap: Record<string, Matter[]> =
    Moon.getSingleton().uploadMattersMap;

  // 抓取模态窗状态
  crawlModalVisible = false;

  // 搜索
  searchState: {
    keyword?: string;
    loading: boolean;
    data: Matter[];
  } = {
    loading: false,
    data: [],
  };

  labelList: Label[] = [];

  // 当前是否有关键词是搜索状态，逻辑上有不一样
  isSearch() {
    return !!this.searchState.keyword;
  }

  //持有全局唯一的实例。
  static instance: List | null = null;

  constructor(props: IProps) {
    super(props);

    List.instance = this;

    this.state = {
      hasSubmission: false,
      tracks: [],
    };
  }

  //拖拽上传
  drag = {
    dragEnterListener: (e: DragEvent) => {
      e.preventDefault();
      this.dragEnterCount++;
      if (this.dragEnterCount > 0) this.updateUI();
    },
    dragleaveListener: (e: DragEvent) => {
      e.preventDefault();
      this.dragEnterCount--;
      if (this.dragEnterCount <= 0) this.updateUI();
    },
    dragoverListener: (e: DragEvent) => {
      e.preventDefault();
    },
    dropListener: (e: DragEvent) => {
      e.preventDefault();
      this.dragEnterCount = 0;
      this.launchUpload(e.dataTransfer!.files);
      this.updateUI();
    },
    register: () => {
      this.wrapperRef.current?.addEventListener(
        'dragenter',
        this.drag.dragEnterListener
      );
      this.wrapperRef.current?.addEventListener(
        'dragleave',
        this.drag.dragleaveListener
      );
      this.wrapperRef.current?.addEventListener(
        'dragover',
        this.drag.dragoverListener
      );
      this.wrapperRef.current?.addEventListener('drop', this.drag.dropListener);
    },
    unregister: () => {
      this.wrapperRef.current?.addEventListener(
        'dragenter',
        this.drag.dragEnterListener
      );
      this.wrapperRef.current?.addEventListener(
        'dragleave',
        this.drag.dragleaveListener
      );
      this.wrapperRef.current?.addEventListener(
        'dragover',
        this.drag.dragoverListener
      );
      this.wrapperRef.current?.addEventListener('drop', this.drag.dropListener);
    },
  };

  paste = {
    pasteListener: (e: ClipboardEvent) => {
      if (e.clipboardData?.files?.length! > 0) {
        this.launchUpload(e.clipboardData?.files!);
      }
    },
    register: () => {
      document.addEventListener('paste', this.paste.pasteListener);
    },
    unregister: () => {
      document.removeEventListener('paste', this.paste.pasteListener);
    },
  };

  async componentDidMount() {
    this.initSpace();
    this.pager.enableHistory();
    this.refresh();
    this.refreshBreadcrumbs();
    this.drag.register();
    this.paste.register();
    let cnt = 0;
    HttpUtil.httpGet(
      '/api/user/label',
      {},
      (res: { data: { data: Label[] } }) => {
        this.labelList = res.data.data.map((l) => {
          l.id = cnt;
          cnt += 1;
          return l;
        });
      }
    );

    // 如果是普通用户，加载赛道信息和检查是否已有提交
    if (this.isRegularUser()) {
      this.loadTracks();
      this.checkSubmission();
    }

  }

  componentWillUnmount() {
    this.drag.unregister();
    this.paste.unregister();
  }

  componentWillReceiveProps() {
    this.pager.enableHistory();
    this.resetSearch();
    this.refresh();
    this.refreshBreadcrumbs();
  }

  initSpace() {
    if (this.props.spaceUuid) {
      this.space.uuid = this.props.spaceUuid;
      this.space.httpDetail(() => this.updateUI());
      this.spaceMember.httpMine(this.props.spaceUuid, () => this.updateUI());
    }
  }

  refresh() {
    // 清空暂存区
    this.selectedMatters = [];

    if (this.isSearch()) {
      // 刷新搜索列表
      console.log('search this2', this.searchState.keyword);
      this.refreshSearch(this.searchState.keyword);
    } else {
      // 刷新分页列表
      this.refreshPager();
    }
    this.updateUI();
  }

  refreshPager() {
    if (!this.pager.getFilterValue('puuid')) {
      this.pager.setFilterValue('puuid', Matter.MATTER_ROOT);
    }

    //如果没有任何的排序，默认使用时间倒序和文件夹在顶部
    if (!this.pager.getCurrentSortFilter()) {
      this.pager.setFilterValue('orderCreateTime', SortDirection.DESC);
      this.pager.setFilterValue('orderDir', SortDirection.DESC);
    }

    // 过滤掉被软删除的文件
    this.pager.setFilterValue('deleted', false);

    // 空间下的文件
    if (this.props.spaceUuid) {
      this.pager.setFilterValue('spaceUuid', this.props.spaceUuid);
    }

    this.pager.httpList();
  }

  getList() {
    if (this.isSearch()) return this.searchState;
    return this.pager;
  }

  getListData() {
    return this.getList().data;
  }

  checkMatter(matter?: Matter) {
    if (matter) {
      if (matter.check) {
        this.selectedMatters.push(matter);
      } else {
        const index = this.selectedMatters.findIndex(
          (item) => item.uuid === matter.uuid
        );
        this.selectedMatters.splice(index, 1);
      }
    } else {
      //统计所有的勾选
      this.selectedMatters = [];
      this.getListData().forEach((matter) => {
        if (matter.check) {
          this.selectedMatters.push(matter);
        }
      });
    }
    this.updateUI();
  }

  checkAll() {
    this.getListData().forEach((i) => {
      i.check = true;
    });
    this.checkMatter();
  }

  checkNone() {
    this.getListData().forEach((i) => {
      i.check = false;
    });
    this.checkMatter();
  }

  deleteBatch() {
    const uuids = this.selectedMatters.map((i) => i.uuid).toString();
    MatterDeleteModal.open(
      () => {
        Matter.httpSoftDeleteBatch(uuids, this.getSpaceUuid(), () => {
          MessageBoxUtil.success(Lang.t('operationSuccess'));
          Capacity.instance?.refresh();
          this.refresh();
        });
      },
      () => {
        Matter.httpDeleteBatch(uuids, this.getSpaceUuid(), () => {
          MessageBoxUtil.success(Lang.t('operationSuccess'));
          Capacity.instance?.refresh();
          this.refresh();
        });
      }
    );
  }

  downloadZip() {
    const uuids = this.selectedMatters.map((i) => i.uuid).toString();
    Matter.downloadZip(uuids, this.getSpaceUuid());
  }

  toggleMoveBatch() {
    MoveBatchModal.open(this.getSpaceUuid(), (targetUuid) => {
      const uuids = this.selectedMatters.map((i) => i.uuid).join(',');
      Matter.httpMove(this.getSpaceUuid(), uuids, targetUuid, () => {
        MessageBoxUtil.success(Lang.t('operationSuccess'));
        this.refresh();
      });
    });
  }

  triggerUpload(fileObj: RcCustomRequestOptions) {
    const { file } = fileObj;
    if (file) this.launchUpload(file as any);
  }

  debounce(func: Function, wait: number) {
    let timer: any = null;
    return (fileObj: any) => {
      const { file } = fileObj;
      this.tempUploadList.push(file);
      if (timer) {
        clearTimeout(timer);
      }
      timer = setTimeout(func, wait);
    };
  }

  triggerUploadDir = this.debounce(async () => {
    await this.uploadDirectory();
    if (!this.uploadErrorLogs.length) {
      MessageBoxUtil.success(Lang.t('matter.uploaded'));
    } else {
      // 上传错误弹出提示框
      const url = FileUtil.getErrorLogsToCSVUrl(this.uploadErrorLogs);
      Modal.confirm({
        title: Lang.t('matter.uploadInfo'),
        icon: <ExclamationCircleFilled twoToneColor="#FFDC00" />,
        content: Lang.t('matter.uploadErrorInfo'),
        okText: Lang.t('matter.exportCSV'),
        onOk: () => {
          if (url) {
            window.open(url);
          } else {
            MessageBoxUtil.warn(Lang.t('matter.unCompatibleBrowser'));
          }
        },
      });
    }

    this.tempUploadList = [];
    this.uploadErrorLogs = [];
    this.setState({ hasSubmission: true });
    this.refresh();
  }, 0);

  async uploadDirectory() {
    const dirPathUuidMap: any = {}; // 存储已经创建好的文件夹map
    for (let i = 0; i < this.tempUploadList.length; i++) {
      const file: File = this.tempUploadList[i];
      const paths = file.webkitRelativePath.split('/');
      const pPaths = paths.slice(0, paths.length - 1);
      const pPathStr = pPaths.join('/');
      if (!dirPathUuidMap[pPathStr]) {
        // 递归创建其父文件夹
        for (let j = 1; j <= pPaths.length; j++) {
          const midPaths = pPaths.slice(0, j);
          const midPathStr = midPaths.join('/');
          if (!dirPathUuidMap[midPathStr]) {
            const m = new Matter(this);
            m.name = midPaths[midPaths.length - 1];
            m.puuid =
              j > 1
                ? dirPathUuidMap[midPaths.slice(0, j - 1).join('/')]
                : this.currentDirectory.uuid!;
            m.spaceUuid = this.getSpaceUuid();
            // 只对根目录文件夹（第一级子文件夹）传递workName和trackId，并标记为rootDirectory
            const isRootDirectory = j === 1 && pPaths.length === 1;
            await m.httpCreateDirectory(
              () => {
                dirPathUuidMap[midPathStr] = m.uuid;
              },
              (msg: string) => {
                this.uploadErrorLogs.push([
                  file.name,
                  file.webkitRelativePath,
                  msg,
                ]);
              },
              undefined,
              isRootDirectory ? this.state.workName : undefined,
              isRootDirectory ? this.state.trackId : undefined,
              isRootDirectory
            );
          }
        }
      }
      if (dirPathUuidMap[pPathStr]) {
        this.launchUpload(file, dirPathUuidMap[pPathStr], (msg?: string) => {
          this.uploadErrorLogs.push([
            file.name,
            file.webkitRelativePath,
            `${msg}`,
          ]);
        });
      }
    }
  }

  launchUpload(
    f: File | FileList,
    puuid = this.currentDirectory.uuid!,
    errHandle = () => {}
  ) {
    const files = f instanceof FileList ? f : [f];
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      // 判断文件是否是文件夹，是的话则停止上传
      const fileReader = new FileReader();
      fileReader.readAsDataURL(files[i].slice(0, 3));
      fileReader.onerror = () => {
        MessageBoxUtil.error(Lang.t('matter.dropNotDirectory'));
      };
      fileReader.onload = () => {
        const m = new Matter(this);
        m.dir = false;
        m.puuid = puuid;
        m.spaceUuid = this.getSpaceUuid();

        //判断文件大小。
        if (this.user.space.sizeLimit >= 0) {
          if (file.size > this.user.space!.sizeLimit) {
            MessageBoxUtil.error(
              Lang.t(
                'matter.sizeExceedLimit',
                StringUtil.humanFileSize(file.size),
                this.user.space?.getHumanizeSizeLimit()
              )
            );
            return;
          }
        }
        m.file = file;
        m.httpUpload(
          () => {
            const index = this.uploadMattersMap[this.getSpaceUuid()].findIndex(
              (matter) => matter === m
            );
            this.uploadMattersMap[this.getSpaceUuid()].splice(index, 1);
            List.instance?.refresh();
            Capacity.instance?.refresh();
          },
          (msg: string) => {
            List.instance?.updateUI();
            SafeUtil.safeCallback(errHandle)(msg);
          }
        );

        if (this.uploadMattersMap[this.getSpaceUuid()]) {
          this.uploadMattersMap[this.getSpaceUuid()].push(m);
        } else {
          this.uploadMattersMap[this.getSpaceUuid()] = [m];
        }
      };
    }
  }

  shareBatch() {
    const uuids = this.selectedMatters.map((i) => i.uuid).join(',');
    ShareOperationModal.open((share: Share) => {
      share.spaceUuid = this.getSpaceUuid();
      // share.httpCreate(uuids, () => {
      //   ShareDialogModal.open(share);
      // });
    });
  }

  changePage(page: number) {
    this.pager.page = page - 1; // page的页数0基
    this.pager.httpList();
    this.updateUI();
  }

  getSpaceUuid() {
    if (this.props.spaceUuid) return this.props.spaceUuid;
    return this.user.spaceUuid!;
  }

  createDirectory() {
    this.newMatter.name = 'matter.allFiles';
    this.newMatter.dir = true;
    this.newMatter.editMode = true;
    this.newMatter.puuid = this.currentDirectory.uuid || Matter.MATTER_ROOT;
    this.newMatter.spaceUuid = this.getSpaceUuid();
    this.director.createMode = true;

    this.updateUI();
    setTimeout(() => this.newMatterRef.current!.highLight());
  }

  previewImage(matter: Matter) {
    let imageArray: string[] = [];
    let startIndex = -1;
    this.getListData().forEach((item) => {
      if (item.isImage()) {
        imageArray.push(item.getPreviewUrl());
        if (item.uuid === matter.uuid) {
          startIndex = imageArray.length - 1;
        }
      }
    });

    ImagePreviewer.showMultiPhoto(imageArray, startIndex);
  }

  goDetail(matter: Matter) {
    const prefix = this.isInSpace ? `/space/${this.getSpaceUuid()}` : '';
    Sun.navigateTo(`${prefix}/matter/detail/${matter.uuid}`);
  }

  pagerAndCapacityRefresh() {
    Capacity.instance?.refresh();
    this.refresh();
  }

  // 如果在空间下，有些操作权限只对管理员和读写成员开放
  checkHandlePermission() {
    if (!this.isInSpace) return true;
    return [SpaceMemberRole.ADMIN, SpaceMemberRole.READ_WRITE].includes(
      this.spaceMember.role!
    );
  }

  handleCloseMatterCrawlModal() {
    this.crawlModalVisible = false;
    this.pagerAndCapacityRefresh();
  }

  goToDirectory(id: string) {
    this.pager.setFilterValue('puuid', id);
    this.pager.page = 0;
    const query = this.pager.getParams();
    Sun.navigateQueryTo({ path: this.getMatterListPath(), query });
  }

  getMatterListPath() {
    if (this.props.spaceUuid)
      return `/space/${this.props.spaceUuid}/matter/list`;
    return '/matter/list';
  }

  // 检查是否是普通用户
  isRegularUser(): boolean {
    return this.user.role === UserRole.USER;
  }

  // 检查是否是评委
  isJudge(): boolean {
    return this.user.role === UserRole.JUDGE;
  }

  // 加载赛道列表
  loadTracks() {
    const track = new Track();
    track.httpList((response: any) => {
      this.setState({ tracks: response.data.data });
    });
  }

  // 检查用户是否已有提交
  checkSubmission() {
    HttpUtil.httpGet('/api/submission/my', {}, (response: any) => {
      this.setState({ hasSubmission: response.data.data !== null });
    });
  }



  // 处理文件夹上传（提交作品）
  handleSubmissionUpload = (
    file: File,
    puuid: string,
    errHandle?: (msg?: string) => void
  ) => {
    if (this.state.hasSubmission) {
      MessageBoxUtil.error('您已经提交过作品，不能重复提交');
      return;
    }

    if (!this.state.trackId) {
      MessageBoxUtil.error('请选择赛道');
      return;
    }

    if (!this.state.workName) {
      MessageBoxUtil.error('请输入作品名称');
      return;
    }

    this.launchUpload(file, puuid, errHandle);
  };

  //刷新面包屑
  refreshBreadcrumbs() {
    this.currentDirectory.uuid =
      this.pager.getFilterValue('puuid') || Matter.MATTER_ROOT;
    const isRootDirectory = this.currentDirectory.uuid === Matter.MATTER_ROOT;

    //根目录简单处理即可。
    const rootBreadcrumb = {
      name: Lang.t('matter.allFiles'),
      path: this.getMatterListPath(),
      query: {},
      displayDirect: isRootDirectory,
    };

    this.breadcrumbModels = [rootBreadcrumb];
    if (!isRootDirectory) {
      this.currentDirectory.spaceUuid = this.getSpaceUuid();
      this.currentDirectory.httpDetail(() => {
        const arr = [];
        let cur: Matter | null = this.currentDirectory;
        do {
          arr.push(cur);
          cur = cur.parent;
        } while (cur);

        this.breadcrumbModels = arr.reduceRight(
          (t: any, item: Matter, i: number) => {
            const query = this.pager.getParams();
            query['puuid'] = item.uuid!;
            t.push({
              name: item.name,
              path: this.getMatterListPath(),
              query: query,
              displayDirect: !i, // 当前目录不需要导航
            });
            return t;
          },
          [rootBreadcrumb]
        );

        this.updateUI();
      });
    }
  }

  getBreadcrumbModels(): BreadcrumbModel[] {
    if (this.props.spaceUuid) {
      if (!this.space.name) return [];
      return [
        {
          name: this.space.name,
          path: '/space',
          query: {},
          displayDirect: false,
        },
      ].concat(this.breadcrumbModels);
    }
    return this.breadcrumbModels;
  }

  getDropdownMenu() {
    return (
      <Menu>
        <Menu.Item
          icon={<CloudUploadOutlined />}
          onClick={() => this.uploadDirectoryBtnRef.current?.click()}
        >
          {Lang.t('matter.uploadDir')}
        </Menu.Item>

        <Menu.Item
          icon={<CloudDownloadOutlined />}
          onClick={() => {
            this.crawlModalVisible = true;
            this.updateUI();
          }}
        >
          {Lang.t('matter.crawl')}
        </Menu.Item>
      </Menu>
    );
  }

  handleChangeSearch = (value?: string) => {
    // 清空暂存区
    this.selectedMatters = [];
    this.searchState.keyword = value;
    this.updateUI();

    if (value) {
      // search
      console.log('search this', value);
      this.refreshSearch(value);
    } else {
      // reset
      this.resetSearch();
    }
  };

  handleTrackChange = (trackId: number) => {
    this.setState({ trackId });
  };

  handleWorkNameChange = (workName: string) => {
    this.setState({ workName });
  };

  refreshSearch = debounce((value: string) => {
    this.searchState.loading = true;
    this.updateUI();
    console.log('this.searchState.keyword', this.searchState.keyword);
    Matter.httpSearch(
      {
        keyword: value,
        puuid: this.currentDirectory.uuid || Matter.MATTER_ROOT,
        spaceUuid: this.getSpaceUuid(),
      },
      (data: Matter[]) => {
        this.searchState.data = data;
      },
      null,
      () => {
        this.searchState.loading = false;
        this.updateUI();
      }
    );
  }, 200);

  resetSearch() {
    this.searchState = {
      keyword: undefined,
      loading: false,
      data: [],
    };
    this.updateUI();
  }

  render() {
    const { pager, director, selectedMatters, dragEnterCount } = this;
    const { hasSubmission, tracks, trackId, workName } = this.state;

    // 如果是普通用户，显示作品提交界面
    const isSubmissionPage = this.isRegularUser() && !this.props.spaceUuid;

    return (
      <div className="matter-list" ref={this.wrapperRef}>
        {dragEnterCount > 0 ? (
          <div className="obscure">
            <CloudUploadOutlined className="white f50" />
          </div>
        ) : null}
        <BreadcrumbPanel breadcrumbModels={this.getBreadcrumbModels()} />

        {isSubmissionPage ? (
          // 作品提交界面
          <div className="submission-panel">
            <h2>作品提交</h2>

            {hasSubmission ? (
              <div className="submission-warning">
                <ExclamationCircleFilled
                  style={{ color: '#faad14', marginRight: 8 }}
                />
                您已经提交过作品，不能重复提交
              </div>
            ) : (
              <>
                <Row gutter={16} className="mb20">
                  <Col xs={24} sm={12}>
                    <div className="form-item">
                      <label>选择赛道：</label>
                      <select
                        className="ant-input"
                        value={trackId || ''}
                        onChange={(e) =>
                          this.handleTrackChange(Number(e.target.value))
                        }
                        disabled={hasSubmission}
                      >
                        <option value="">请选择赛道</option>
                        {tracks.map((track) => (
                          <option key={track.id} value={track.id}>
                            {track.name}
                          </option>
                        ))}
                      </select>
                    </div>
                  </Col>
                  <Col xs={24} sm={12}>
                    <div className="form-item">
                      <label>作品名称：</label>
                      <Input
                        value={workName || ''}
                        onChange={(e) =>
                          this.handleWorkNameChange(e.target.value)
                        }
                        placeholder="请输入作品名称"
                        disabled={hasSubmission}
                      />
                    </div>
                  </Col>
                </Row>

                <div className="upload-section">
                  <h3>上传作品文件夹</h3>
                  <p className="upload-hint">
                    请将您的作品文件整理到一个文件夹中，然后上传整个文件夹
                  </p>

                  <Upload
                    className="ant-upload "
                    customRequest={this.triggerUploadDir}
                    showUploadList={false}
                    directory
                  >
                    <Button
                      type="primary"
                      icon={<CloudUploadOutlined />}
                      disabled={hasSubmission || !trackId || !workName}
                    >
                      {hasSubmission ? '已提交作品' : '选择文件夹并上传'}
                    </Button>
                  </Upload>
                </div>
              </>
            )}
          </div>
        ) : (
          // 原来的文件管理界面
          <Row>
            <Col xs={24} sm={24} md={14} lg={16} className="mt10">
              <AntdSpace className="buttons">
                {selectedMatters.length !== this.getListData().length ? (
                  <Button
                    type="primary"
                    className="mb10"
                    onClick={() => this.checkAll()}
                  >
                    <PlusSquareOutlined />
                    {Lang.t('selectAll')}
                  </Button>
                ) : null}
                {this.getListData().length &&
                selectedMatters.length === this.getListData().length ? (
                  <Button
                    type="primary"
                    className="mb10"
                    onClick={() => this.checkNone()}
                  >
                    <MinusSquareOutlined />
                    {Lang.t('cancel')}
                  </Button>
                ) : null}
                {selectedMatters.length ? (
                  <>
                    {!this.isSearch() && (
                      <Button
                        type="primary"
                        className="mb10"
                        onClick={() => this.downloadZip()}
                      >
                        <DownloadOutlined />
                        {Lang.t('download')}
                      </Button>
                    )}
                    {this.checkHandlePermission() && (
                      <>
                        <Button
                          type="primary"
                          className="mb10"
                          onClick={() => this.deleteBatch()}
                        >
                          <DeleteOutlined />
                          {Lang.t('delete')}
                        </Button>
                        <Button
                          type="primary"
                          className="mb10"
                          onClick={() => this.toggleMoveBatch()}
                        >
                          <DragOutlined />
                          {Lang.t('matter.move')}
                        </Button>
                      </>
                    )}
                  </>
                ) : null}

                {this.checkHandlePermission() && !this.isSearch() && (
                  <>
                    <Upload
                      className="ant-upload"
                      customRequest={(e) => this.triggerUpload(e)}
                      showUploadList={false}
                      multiple
                    >
                      <Button type="primary" className="mb10">
                        <CloudUploadOutlined />
                        {Lang.t('matter.upload')}
                      </Button>
                    </Upload>
                    <Button
                      type="primary"
                      className="mb10"
                      onClick={() =>
                        this.uploadDirectoryBtnRef.current?.click()
                      }
                    >
                      <CloudUploadOutlined />
                      {Lang.t('matter.uploadDir')}
                    </Button>
                    <Button
                      type="primary"
                      className="mb10"
                      onClick={() => this.createDirectory()}
                    >
                      <FolderOutlined />
                      {Lang.t('matter.create')}
                    </Button>
                  </>
                )}

                <Button
                  type="primary"
                  className="mb10"
                  onClick={() => this.refresh()}
                >
                  <SyncOutlined />
                  {Lang.t('refresh')}
                </Button>

                <Upload
                  className="ant-upload display-none"
                  customRequest={this.triggerUploadDir}
                  showUploadList={false}
                  directory
                >
                  <Button type="primary" ref={this.uploadDirectoryBtnRef}>
                    <CloudUploadOutlined />
                    {Lang.t('matter.uploadDir')}
                  </Button>
                </Upload>
              </AntdSpace>
            </Col>
            <Col xs={24} sm={24} md={10} lg={8} className="mt10">
              <Input
                className="mb10"
                placeholder={Lang.t('matter.searchFile')}
                prefix={<SearchOutlined />}
                value={this.searchState.keyword}
                onChange={(e) => this.handleChangeSearch(e.target.value)}
                allowClear
              />
            </Col>
          </Row>
        )}

        {Children.toArray(
          this.uploadMattersMap[this.getSpaceUuid()]?.map((m) => (
            <UploadMatterPanel key={m.autoId} matter={m} />
          ))
        )}

        {this.getListData().length ? (
          <MatterSortPanel
            pager={pager}
            refresh={() => this.refresh()}
            disabled={this.isSearch()}
          />
        ) : null}

        {director.createMode ? (
          <MatterPanel
            mode="normal"
            ref={this.newMatterRef}
            matter={this.newMatter}
            allLabels={this.labelList}
            userUuid={this.user.uuid ?? ''}
            director={director}
            onCreateDirectoryCallback={() => this.refresh()}
          />
        ) : null}
        <div>
          {this.getList().loading || this.getList().data.length > 0 ? (
            <AntdList
              rowKey="uuid"
              className="matter-list-content"
              loading={this.getList().loading}
              dataSource={this.getList().data}
              renderItem={(matter) => (
                <MatterPanel
                  mode={this.isInSpace ? 'space' : 'normal'}
                  spaceMemberRole={
                    this.isInSpace ? this.spaceMember.role! : undefined
                  }
                  key={matter.uuid}
                  director={director}
                  userUuid={this.user.uuid ?? ''}
                  matter={matter}
                  allLabels={this.labelList}
                  onGoToDirectory={(id) => this.goToDirectory(id)}
                  onDeleteSuccess={() => this.pagerAndCapacityRefresh()}
                  onCheckMatter={(m) => this.checkMatter(m)}
                  onPreviewImage={(m) => this.previewImage(m)}
                  onGoDetail={(m) => this.goDetail(m)}
                />
              )}
              pagination={
                this.isSearch()
                  ? false
                  : {
                      onChange: (page) => this.changePage(page),
                      current: pager.page + 1,
                      total: pager.totalItems,
                      pageSize: pager.pageSize,
                      hideOnSinglePage: true,
                    }
              }
            />
          ) : (
            <Empty description={Lang.t('matter.noContentYet')} />
          )}
        </div>
        {this.crawlModalVisible && (
          <MatterCrawlModal
            spaceUuid={this.getSpaceUuid()}
            puuid={this.currentDirectory.uuid!}
            onClose={this.handleCloseMatterCrawlModal.bind(this)}
          />
        )}
      </div>
    );
  }
}

import BaseEntity from '../base/BaseEntity';
import SafeUtil from '../../util/SafeUtil';

export default class Track extends BaseEntity {
  static URL_API_TRACK_LIST = '/api/track/list';

  id: number = 0;
  name: string = '';
  targetUserType: string = 'BOTH';
  description: string = '';

  tracks: any[] = [];

  constructor(reactComponent?: React.Component) {
    super(reactComponent);
  }

  assign(obj: any) {
    super.assign(obj);
  }

  getForm(): any {
    return {
      id: this.id,
      name: this.name,
      targetUserType: this.targetUserType,
      description: this.description,
    };
  }

  httpList(successCallback?: any, errorCallback?: any, finalCallback?: any) {
    let that = this;
    this.httpGet(
      Track.URL_API_TRACK_LIST,
      {},
      function (response: any) {
        that.tracks = response.data.data;
        SafeUtil.safeCallback(successCallback)(response);
      },
      errorCallback,
      finalCallback
    );
  }

  httpCreate(name: string, targetUserType: string, description: string, successCallback?: any, errorCallback?: any, finalCallback?: any) {
    let that = this;
    this.httpPost(
      '/api/track/create',
      { name, targetUserType, description },
      function (response: any) {
        SafeUtil.safeCallback(successCallback)(response);
      },
      errorCallback,
      finalCallback
    );
  }

  httpDelete(id: number, successCallback?: any, errorCallback?: any, finalCallback?: any) {
    let that = this;
    this.httpPost(
      '/api/track/delete',
      { id },
      function (response: any) {
        SafeUtil.safeCallback(successCallback)(response);
      },
      errorCallback,
      finalCallback
    );
  }
}
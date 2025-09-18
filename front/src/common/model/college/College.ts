import BaseEntity from '../base/BaseEntity';
import SafeUtil from '../../util/SafeUtil';

export default class College extends BaseEntity {
  static URL_API_COLLEGE_LIST = '/api/college/list';

  id: number = 0;
  name: string = '';

  colleges: any[] = [];

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
    };
  }

  httpList(successCallback?: any, errorCallback?: any, finalCallback?: any) {
    let that = this;
    this.httpGet(
      College.URL_API_COLLEGE_LIST,
      {},
      function (response: any) {
        that.colleges = response.data.data;
        SafeUtil.safeCallback(successCallback)(response);
      },
      errorCallback,
      finalCallback
    );
  }

  httpCreate(name: string, successCallback?: any, errorCallback?: any, finalCallback?: any) {
    let that = this;
    this.httpPost(
      '/api/college/create',
      { name },
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
      '/api/college/delete',
      { id },
      function (response: any) {
        SafeUtil.safeCallback(successCallback)(response);
      },
      errorCallback,
      finalCallback
    );
  }

  httpBulkCreate(names: string, successCallback?: any, errorCallback?: any, finalCallback?: any) {
    let that = this;
    this.httpPost(
      '/api/college/bulk-create',
      { names },
      function (response: any) {
        SafeUtil.safeCallback(successCallback)(response);
      },
      errorCallback,
      finalCallback
    );
  }
}
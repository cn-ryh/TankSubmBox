import BaseEntity from '../../base/BaseEntity';

export default class CollegeConfig extends BaseEntity {
  enabledColleges: number[] = [];

  constructor(reactComponent?: React.Component) {
    super(reactComponent);
  }

  assign(obj: any) {
    super.assign(obj);
  }

  getForm(): any {
    return {
      enabledColleges: this.enabledColleges,
    };
  }
}
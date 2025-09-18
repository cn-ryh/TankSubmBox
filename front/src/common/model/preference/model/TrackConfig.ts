import BaseEntity from '../../base/BaseEntity';

export default class TrackConfig extends BaseEntity {
  enabledTracks: number[] = [];

  constructor(reactComponent?: React.Component) {
    super(reactComponent);
  }

  assign(obj: any) {
    super.assign(obj);
  }

  getForm(): any {
    return {
      enabledTracks: this.enabledTracks,
    };
  }
}
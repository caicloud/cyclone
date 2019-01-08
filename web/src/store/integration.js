import { observable, action } from 'mobx';

// useStrict(true);

class Integration {
  @observable integrationList = [];
  @action.bound
  getIntegrationList(workflowId) {
    this.integrationList = [
      {
        name: 'svn-2',
        type: 'svn-1',
        time: '2018-12-21',
      },
    ];
    return;
  }
}

export default new Integration();

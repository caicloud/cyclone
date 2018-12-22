import { observable, useStrict, action } from 'mobx';
// import fetchApi from "../api/index.js";

useStrict(true);

class Integration {
  @observable integrationList = [];
  @action.bound
  getIntegrationList(workflowId) {
    this.integrationList = [
      {
        id: '5c04e7a73c17eb00019e5a32',
        name: 'svn-2',
        type: 'svn-1',
        time: '2018-12-21',
      },
    ];
    return;
  }
}

export default new Integration();

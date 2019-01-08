import { observable, action } from 'mobx';
// import fetchApi from "../api/index.js";

// useStrict(true);

class Workflow {
  @observable workflowList = [];
  @action.bound
  getWorkflowList(workflowId) {
    this.workflowList = [
      {
        id: '5c04e7a73c17eb00019e5a32',
        name: 'svn-1',
        alias: 'svn-1',
        owner: 'admin',
        recentVersion: 'v1.1.0',
        creationTime: '2018-09-09',
        projectID: '5c04dcef3c17eb00019e5a2d',
      },
    ];
    // TODO
    // return fetchApi.fetchWorkflowList("svn-trigger", {}).then(data => {

    // });
    return;
  }
}

export default new Workflow();

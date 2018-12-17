import { observable, useStrict, action } from 'mobx';
// import fetchApi from "../api/index.js";

useStrict(true);

class Pipeline {
  @observable pipelineList = [];
  @action.bound
  getPipelineList(pipelineId) {
    this.pipelineList = [
      {
        id: '5c04e7a73c17eb00019e5a32',
        name: 'svn-1',
        alias: 'svn-1',
        owner: 'admin',
        projectID: '5c04dcef3c17eb00019e5a2d',
      },
    ];
    // TODO
    // return fetchApi.fetchPipelineList("svn-trigger", {}).then(data => {

    // });
    return;
  }
}

export default new Pipeline();

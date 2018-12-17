import http from './http.js';

console.log('instance', http);

const fetchApi = {
  http,
  fetchPipelineList(workspace, query) {
    return http.get(`/workspaces/${workspace}/pipelines`, query).then(data => {
      return data;
    });
  },
};

export default fetchApi;

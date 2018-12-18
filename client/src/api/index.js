import http from './http.js';

const fetchApi = {
  http,
  fetchPipelineList(workspace, query) {
    return http.get(`/workspaces/${workspace}/pipelines`, query).then(data => {
      return data;
    });
  },
};

export default fetchApi;

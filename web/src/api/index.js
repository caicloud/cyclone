import http from './http.js';

const fetchApi = {
  http,
  fetchWorkflowList(project, query) {
    return http.get(`/projects/${project}/workflows`, query).then(data => {
      return data;
    });
  },
};

export default fetchApi;

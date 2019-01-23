import http from './http.js';

const fetchApi = {
  http,
  fetchWorkflowList(project, query) {
    return http.get(`/projects/${project}/workflows`, query).then(data => {
      return data;
    });
  },
  fetchStageTemplates(query) {
    return http.get(`/stages/templates`, query).then(data => {
      return data;
    });
  },
  fetchIntegrationList(query) {
    return http.get(`/integrations`, query).then(data => {
      return data;
    });
  },
};

export default fetchApi;

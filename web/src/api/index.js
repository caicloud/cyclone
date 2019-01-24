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
  fetchResources() {
    return http.get('/resources').then(data => {
      return data;
    });
  },
  /** start project */
  listProjects() {
    return http.get('/projects').then(data => {
      return data;
    });
  },
  createProject(data) {
    return http.post('/projects', data);
  },
  removeProject(name) {
    return http.delete(`/projects/${name}`);
  },
  getProject(name) {
    return http.get(`/projects/${name}`);
  },
  /** end project */
};

export default fetchApi;

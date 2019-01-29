import http from './http.js';

const fetchApi = {
  http,
  fetchWorkflowList(project, query) {
    return http.get(`/projects/${project}/workflows`, query).then(data => {
      return data;
    });
  },
  fetchStageTemplates(query) {
    return http.get(`/templates`, query).then(data => {
      return data;
    });
  },
  fetchIntegrationList(query) {
    return http.get(`/integrations`, query).then(data => {
      return data;
    });
  },
  createIntegration(data) {
    return http.post('/integrations', data);
  },
  removeIntegration(name) {
    return http.delete(`/integrations/${name}`);
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
  updateProject(data, name) {
    return http.put(`/projects/${name}`, data);
  },
  removeProject(name) {
    return http.delete(`/projects/${name}`);
  },
  getProject(name) {
    return http.get(`/projects/${name}`);
  },
  listProjectResources(project) {
    return http.get(`/projects/${project}/resources`).then(data => {
      return data;
    });
  },
  listProjectStages(project) {
    return http.get(`/projects/${project}/stages`).then(data => {
      return data;
    });
  },
  /** end project */
};

export default fetchApi;

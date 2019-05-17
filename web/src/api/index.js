import http from './http.js';

const fetchApi = {
  http,
  listWorkflow(project, query) {
    return http.get(`/projects/${project}/workflows`, query).then(data => {
      return data;
    });
  },
  // start template
  fetchStageTemplates(query) {
    return http.get('/templates', query).then(data => {
      return data;
    });
  },
  getStageTemplate(template) {
    return http.get(`/templates/${template}`).then(data => data);
  },
  createStageTemplate(data) {
    return http.post('/templates', data);
  },
  updateStageTemplate(data, name) {
    return http.put(`/templates/${name}`, data);
  },
  removeStageTemplate(name) {
    return http.delete(`/templates/${name}`);
  },
  // end template
  fetchIntegrationList(query) {
    return http.get(`/integrations`, query).then(data => {
      return data;
    });
  },
  createIntegration(data) {
    return http.post('/integrations', data);
  },
  updateIntegration(data, name) {
    return http.put(`/integrations/${name}`, data);
  },
  getIntegration(name) {
    return http.get(`/integrations/${name}`);
  },
  removeIntegration(name) {
    return http.delete(`/integrations/${name}`);
  },
  openCluster(name) {
    return http.put(`/integrations/${name}/opencluster`, null);
  },
  closeCluster(name) {
    return http.put(`/integrations/${name}/closecluster`, null);
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
  /** start resource */
  createResource(project, info) {
    return http.post(`/projects/${project}/resources`, info).then(data => {
      return data;
    });
  },

  createStage(project, info) {
    return http.post(`/projects/${project}/stages`, info).then(data => {
      return data;
    });
  },
  /** end resource */
  createWorkflow(project, info) {
    return http.post(`/projects/${project}/workflows`, info).then(data => {
      return data;
    });
  },

  removeWorkflow(project, workflow) {
    return http
      .delete(`/projects/${project}/workflows/${workflow}`)
      .then(data => {
        return data;
      });
  },

  getWorkflow(project, workflow) {
    return http.get(`/projects/${project}/workflows/${workflow}`).then(data => {
      return data;
    });
  },
};

export default fetchApi;

import http from './http.js';

const fetchApi = {
  http,
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
    return http.get(`/integrations`, { params: query }).then(data => {
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
  listProjects(query = {}) {
    return http.get('/projects', { params: query }).then(data => {
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

  getResource(project, resource) {
    return http.get(`/projects/${project}/resources/${resource}`).then(data => {
      return data;
    });
  },

  createResource(project, info) {
    return http.post(`/projects/${project}/resources`, info).then(data => {
      return data;
    });
  },

  updateResource(project, resource, info) {
    return http
      .put(`/projects/${project}/resources/${resource}`, info)
      .then(data => {
        return data;
      });
  },

  deleteResource(project, resource) {
    return http
      .delete(`/projects/${project}/resources/${resource}`)
      .then(data => {
        return data;
      });
  },

  createStage(project, info) {
    return http.post(`/projects/${project}/stages`, info).then(data => {
      return data;
    });
  },

  getStage(project, stage) {
    return http.get(`/projects/${project}/stages/${stage}`).then(data => {
      return data;
    });
  },

  updateStage(project, stage, info) {
    return http.put(`/projects/${project}/stages/${stage}`, info).then(data => {
      return data;
    });
  },

  deleteStage(project, stage) {
    return http.delete(`/projects/${project}/stages/${stage}`).then(data => {
      return data;
    });
  },

  listResourceTypes(query) {
    return http.get('/resourcetypes', { params: query }).then(data => {
      return data;
    });
  },

  getResourceType(type) {
    return http.get(`/resourcetypes/${type}`).then(data => {
      return data;
    });
  },

  createResourceType(data) {
    return http.post(`/resourcetypes`, data).then(data => {
      return data;
    });
  },

  updateResourceType(resourceType, data) {
    return http.put(`/resourcetypes/${resourceType}`, data).then(data => {
      return data;
    });
  },

  deleteResourceType(resourceType) {
    return http.delete(`/resourcetypes/${resourceType}`, null).then(data => {
      return data;
    });
  },

  listWorkflow(project, query = {}) {
    return http
      .get(`/projects/${project}/workflows`, { params: query })
      .then(data => {
        return data;
      });
  },

  createWorkflow(project, info) {
    return http.post(`/projects/${project}/workflows`, info).then(data => {
      return data;
    });
  },

  updateWorkflow(project, workflow, info) {
    return http
      .put(`/projects/${project}/workflows/${workflow}`, info)
      .then(data => {
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

  listSCMRepos(integration) {
    return http.get(`/integrations/${integration}/scmrepos`).then(data => {
      return data;
    });
  },

  runWorkflow(projectName, workflowName, info) {
    return http
      .post(
        `/projects/${projectName}/workflows/${workflowName}/workflowruns`,
        info
      )
      .then(data => {
        return data;
      });
  },

  listWorkflowRuns(project, workflow, query) {
    return http
      .get(`/projects/${project}/workflows/${workflow}/workflowruns`, {
        params: query,
      })
      .then(data => {
        return data;
      });
  },

  deleteWorkflowRun(project, workflow, record) {
    return http
      .delete(
        `/projects/${project}/workflows/${workflow}/workflowruns/${record}`
      )
      .then(data => {
        return data;
      });
  },

  getStorageUsage() {
    return http.get('/storage/usages').then(data => data);
  },

  listAllWorkflows() {
    return http.get('/workflows').then(data => {
      return data;
    });
  },

  listAllWorkflowRuns() {
    return http.get('/workflowruns').then(data => {
      return data;
    });
  },

  getWorkflowrun(project, workflow, workflowRun, config = {}) {
    return http
      .get(
        `/projects/${project}/workflows/${workflow}/workflowruns/${workflowRun}`,
        config
      )
      .then(data => {
        return data;
      });
  },

  getWorkflowRunLog(params, query) {
    return http
      .get(
        `/projects/${_.get(params, 'projectName')}/workflows/${_.get(
          params,
          'workflowName'
        )}/workflowruns/${_.get(params, 'workflowRun')}/logs?container=${
          query.container
        }&stage=${query.stage}`
      )
      .then(data => {
        return data;
      });
  },
};

export default fetchApi;

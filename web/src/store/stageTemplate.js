import { observable, action } from 'mobx';
import fetchApi from '../api/index.js';

class StageTemplate {
  @observable template = {};
  @observable templateLoading = false;
  @observable templateList = [];
  @observable templateListLoading = false;

  @action
  getTemplate(templateName) {
    this.templateLoading = true;
    fetchApi.getStageTemplate(templateName).then(this.getTemplateSuccess);
  }

  @action
  createStageTemplate(data, cb) {
    fetchApi.createStageTemplate(data).then(() => {
      cb();
    });
  }

  @action
  updateStageTemplate(data, name, cb) {
    fetchApi.updateStageTemplate(data, name).then(() => {
      cb();
    });
  }

  @action.bound
  deleteStageTemplate(name, cb) {
    fetchApi.removeStageTemplate(name).then(() => {
      cb();
    });
  }

  @action
  getTemplateList(cb) {
    this.templateList = [];
    this.templateListLoading = true;
    fetchApi.fetchStageTemplates({}).then(data => {
      this.fetchStageTemplatesSuccess(data);
      cb && cb(data);
    });
  }

  @action.bound
  getTemplateSuccess(template) {
    this.templateLoading = false;
    this.template = template;
  }

  @action.bound
  fetchStageTemplatesSuccess(templateList) {
    this.templateListLoading = false;
    this.templateList = templateList;
  }
  @action.bound
  resetTemplate() {
    this.template = null;
  }
}

export default new StageTemplate();

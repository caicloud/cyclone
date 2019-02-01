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
  getTemplateList() {
    this.templateList = [];
    this.templateListLoading = true;
    fetchApi.fetchStageTemplates({}).then(this.fetchStageTemplatesSuccess);
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
}

export default new StageTemplate();

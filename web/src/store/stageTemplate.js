import { observable, useStrict, action } from 'mobx';
import fetchApi from '../api/index.js';

useStrict(true);

class StageTemplate {
  @observable templateList = [];
  @observable templateListLoading = false;

  @action
  getTemplateList() {
    this.templateList = [];
    this.templateListLoading = true;
    fetchApi.fetchStageTemplates({}).then(this.fetchStageTemplatesSuccess);
  }

  @action.bound
  fetchStageTemplatesSuccess(templateList) {
    this.templateListLoading = false;
    this.templateList = templateList;
  }
}

export default new StageTemplate();

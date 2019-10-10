import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { Spin } from 'antd';
import CreateWorkflow from './CreateWorkflow';
import { inject, observer } from 'mobx-react';
import { formatSubmitData, revertWorkflow } from './util';
import fetchApi from '@/api/index.js';

@inject('resource', 'workflow')
@observer
class AddWorkflow extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      depend: {},
      submitting: false,
      position: {},
    };
    const {
      match: {
        params: { workflowName, projectName },
      },
      workflow,
    } = props;
    if (workflowName) {
      workflow.getWorkflow(projectName, workflowName);
    }
  }

  setStageDepend = (targetID, sourceName, targeName) => {
    const { depend } = this.state;
    const obj = _.cloneDeep(depend);
    if (sourceName) {
      if (!obj[targetID]) {
        obj[targetID] = [sourceName];
      } else if (obj[targetID].includes(sourceName)) {
        const index = _.findIndex(obj[targetID], o => o === sourceName);
        _.pullAt(obj[targetID], index);
      } else {
        obj[targetID].push(sourceName);
      }
    } else {
      delete obj[targetID];
      _.forEach(obj, v => {
        if (v && v.includes(targeName)) {
          const index = v.indexOf(targeName);
          _.pullAt(v, index);
        }
      });
    }

    this.setState({ depend: obj });
  };

  saveStagePosition = (stageId, data) => {
    const { position } = this.state;
    const _position = _.cloneDeep(position);
    if (!_.isEmpty(data)) {
      _position[stageId] = data;
    } else {
      delete _position[stageId];
    }
    this.setState({ position: _position });
  };

  // submit form data
  submit = value => {
    const {
      match: {
        params: { projectName },
      },
    } = this.props;
    this.setState({ submitting: true });
    const requests = formatSubmitData(value, projectName, this.state);
    this.postAllRequests(requests).then(data => {
      this.setState({ submitting: false });
      if (!_.get(data, 'submitError')) {
        this.props.history.push(`/projects/${projectName}`);
      }
    });
  };

  async postAllRequests(requests) {
    for (const req of requests) {
      try {
        const fn =
          req.type === 'createWorkflow'
            ? fetchApi.createWorkflow
            : req.type === 'createStage'
            ? fetchApi.createStage
            : fetchApi.createResource;
        await fn(req.project, req.data);
      } catch (err) {
        return { submitError: true };
      }
    }
  }

  getInitialValues = data => {
    let defaultValue = {
      metadata: { name: '', annotations: { description: '' } },
      stages: [],
      currentStage: '',
    };

    if (!_.isEmpty(data)) {
      defaultValue = revertWorkflow(data);
    }
    return defaultValue;
  };

  render() {
    const {
      match: { params },
      workflow: { workflowDetail },
    } = this.props;
    const { submitting } = this.state;
    if (
      _.get(params, 'workflowName') &&
      !_.get(workflowDetail, `${params.workflowName}`)
    ) {
      return <Spin />;
    }

    const initValue = this.getInitialValues(
      _.get(workflowDetail, `${params.workflowName}`)
    );
    return (
      <Formik
        initialValues={initValue}
        onSubmit={this.submit}
        render={props => (
          <CreateWorkflow
            {...props}
            project={params.projectName}
            workflowName={_.get(params, 'workflowName')}
            workFlowInfo={_.get(workflowDetail, `${params.workflowName}`)}
            handleDepend={this.setStageDepend}
            saveStagePostition={this.saveStagePosition}
            submitting={submitting}
            history={this.props.history}
          />
        )}
      />
    );
  }
}

AddWorkflow.propTypes = {
  handleSubmit: PropTypes.func,
  resource: PropTypes.object,
  workflow: PropTypes.object,
  history: PropTypes.object,
  match: PropTypes.object,
};

export default AddWorkflow;

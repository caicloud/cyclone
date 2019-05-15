import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { Spin } from 'antd';
import CreateWorkflow from './CreateWorkflow';
import { inject, observer } from 'mobx-react';
import { formatSubmitData, revertWorkflow } from './util';
import { getQuery } from '@/lib/util';
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
      match: { params },
      history: { location },
      workflow,
    } = props;
    if (_.get(params, 'workflowName')) {
      const query = getQuery(location.search);
      workflow.getWorkflow(query.project, params.workflowName);
    }
  }

  setStageDepend = (target, sourceName) => {
    const { depend } = this.state;
    const obj = _.cloneDeep(depend);
    if (!obj[target]) {
      obj[target] = [sourceName];
    } else if (obj[target].includes(sourceName)) {
      const index = _.findIndex(obj[target], o => o === sourceName);
      _.pullAt(obj[target], index);
    } else {
      obj[target].push(sourceName);
    }
    this.setState({ depend: obj });
  };

  saveStagePosition = (stageId, data) => {
    const { position } = this.state;
    const _position = _.cloneDeep(position);
    _position[stageId] = data;
    this.setState({ position: _position });
  };

  // submit form data
  submit = value => {
    const {
      history: { location },
    } = this.props;
    console.log('before submit', JSON.stringify(value));
    this.setState({ submitting: true });
    const query = getQuery(location.search);
    const requests = formatSubmitData(value, query, this.state);
    this.postAllRequests(requests).then(data => {
      this.setState({ submitting: false });
      if (!_.get(data, 'submitError')) {
        this.props.history.push(`/workflow?project=${query.project}`);
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

  validate = () => {
    const errors = {};
    return errors;
  };

  getInitialValues = data => {
    let defaultValue = {
      metadata: { name: '', description: '' },
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
      history: { location },
      workflow: { workflowDetail },
    } = this.props;
    const { submitting } = this.state;
    const query = getQuery(location.search);
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
        validate={this.validate}
        onSubmit={this.submit}
        render={props => (
          <CreateWorkflow
            {...props}
            project={query.project}
            workFlowInfo={_.get(workflowDetail, `${params.workflowName}`)}
            handleDepend={this.setStageDepend}
            saveStagePostition={this.saveStagePosition}
            submitting={submitting}
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
};

export default AddWorkflow;

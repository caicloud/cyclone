import PropTypes from 'prop-types';
import { Formik } from 'formik';
import CreateWorkflow from './CreateWorkflow';
import { inject, observer } from 'mobx-react';
import qs from 'query-string';

@inject('resource', 'workflow')
@observer
class AddWorkflow extends React.Component {
  state = {
    depend: {},
  };
  setStageDepend = depend => {
    const obj = {};
    _.forEach(depend, v => {
      if (!obj[v.target]) {
        obj[v.target] = [v.source];
      } else {
        obj[v.target].push(v.source);
      }
    });
    this.setState({ depend: obj });
  };

  tramformArg = data => {
    const value = _.cloneDeep(data);
    const containers = _.get(value, 'spec.containers[0]');
    _.forEach(containers, (v, k) => {
      if (['args', 'command'].includes(k) && _.isString(v)) {
        value.spec.containers[0][k] = v.split(/(?:\r\n|\r|\n)/);
      }
    });
    return value;
  };

  // submit form data
  submit = value => {
    const {
      history: { location },
    } = this.props;
    const { depend } = this.state;
    const requests = [];
    const query = qs.parse(location.search);
    const stages = _.get(value, 'stages', []);
    const workflowInfo = {
      metadata: _.get(value, 'metadata'),
      spec: {
        stages: [],
      },
    };
    _.forEach(stages, v => {
      const inputResources = _.get(value, `${v}.inputs.resources`, []);
      const outputResources = _.get(value, `${v}.outputs.resources`, []);
      let stage = {
        metadata: {
          name: _.get(value, `${v}.name`),
        },
        spec: {
          pod: this.tramformArg(
            _.pick(_.get(value, v), ['inputs', 'outputs', 'spec'])
          ),
        },
      };
      _.forEach(inputResources, (r, i) => {
        const data = _.pick(r, ['spec']);
        data.metadata = {
          name: r.name,
        };
        stage.spec.pod.inputs.resources[i] = _.pick(r, ['name', 'path']);
        requests.push({ type: 'createResource', project: query.project, data });
      });

      _.forEach(outputResources, (r, i) => {
        const data = _.pick(r, ['spec']);
        data.metadata = {
          name: r.name,
        };
        stage.spec.pod.outputs.resources[i] = _.pick(r, ['name']);
        requests.push({ type: 'createResource', project: query.project, data });
      });
      requests.push({
        type: 'createStage',
        project: query.project,
        data: stage,
      });

      const workflowStage = {
        artifacts: _.get(v, 'outputs.artifacts', []),
        depend: _.get(depend, v),
        name: _.get(value, `${v}.name`),
      };
      workflowInfo.spec.stages.push(workflowStage);
    });
    requests.push({
      type: 'createWorkflow',
      project: query.project,
      data: workflowInfo,
    });

    this.postAllRequests(requests);
  };

  async postAllRequests(requests) {
    const {
      resource: { createStage, createResource },
      workflow: { createWorkflow },
    } = this.props;
    for (const req of requests) {
      const fn =
        req.type === 'createWorkflow'
          ? createWorkflow
          : req.type === 'createStage'
          ? createStage
          : createResource;
      await fn(req.project, req.data);
    }
    // TODO(qme): catch error
  }

  validate = () => {
    const errors = {};
    return errors;
  };

  getInitialValues = () => {
    let defaultValue = {
      metadata: { name: '', description: '' },
      stages: [],
      currentStage: '',
    };
    return defaultValue;
  };

  render() {
    const initValue = this.getInitialValues();
    return (
      <Formik
        initialValues={initValue}
        validate={this.validate}
        onSubmit={this.submit}
        render={props => (
          <CreateWorkflow {...props} handleDepend={this.setStageDepend} />
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

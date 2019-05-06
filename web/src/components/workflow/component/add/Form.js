import PropTypes from 'prop-types';
import { Formik } from 'formik';
import CreateWorkflow from './CreateWorkflow';
import { inject, observer } from 'mobx-react';

@inject('integration')
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

  // submit form data
  submit = value => {
    const { integration } = this.props;
    const { depend } = this.state;
    const stages = _.get(value, 'stages', []);
    const workflow = {
      spec: {
        stages: [],
      },
    };
    _.forEach(stages, v => {
      const resources = _.pick(value, `${v}.inputs.resources`);
      _.forEach(resources, r => {
        const data = _.pick(r, 'spec');
        data.metadata = {
          name: r.name,
        };
        integration.createIntegration(data);
        const stage = {
          artifacts: _.get(v, 'outputs.artifacts', []),
          depend: _.get(depend, `${r.name}`),
          name: v.metadata.name,
        };
        workflow.spec.stages.push(stage);
      });
    });
  };

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
  integration: PropTypes.object,
};

export default AddWorkflow;

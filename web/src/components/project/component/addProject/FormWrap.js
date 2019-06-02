import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { toJS } from 'mobx';
import { Spin } from 'antd';
import FormContent from './FormContent';
import { inject, observer } from 'mobx-react';

@inject('project')
@observer
class AddProject extends React.Component {
  constructor(props) {
    super(props);
    const {
      match: { params },
    } = props;
    const update = !!_.get(params, 'projectName');
    this.state = {
      update,
    };
    if (update) {
      props.project.getProject(params.projectName);
    }
  }

  // submit form data
  submit = values => {
    const {
      project,
      history,
      match: { params },
    } = this.props;
    const { update } = this.state;
    const data = _.omit(values, 'metadata');
    data.metadata = {
      annotations: {
        'cyclone.dev/description': _.get(values, 'metadata.description', ''),
        'cyclone.dev/alias': _.get(values, 'metadata.alias', ''),
      },
    };
    if (update) {
      project.updateProject(data, params.projectName, () => {
        history.replace(`/project`);
      });
    } else {
      project.createProject(data, () => {
        history.replace(`/project`);
      });
    }
  };

  validate = values => {
    const errors = {};
    if (!values.metadata.alias) {
      errors.metadata = { alias: intl.get('validate.required') };
    }

    return errors;
  };

  getInitialValues = () => {
    const { update } = this.state;
    let defaultValue = {
      metadata: { alias: '', description: '' },
      spec: {
        quota: {
          'limits.cpu': '',
          'limits.memory': '',
          'requests.cpu': '',
          'requests.memory': '',
        },
      },
    };
    if (update) {
      const proejctInfo = toJS(this.props.project.projectDetail);
      const values = _.pick(proejctInfo, ['spec.quota']);
      const metadata = _.get(proejctInfo, ['metadata', 'annotations']);
      values.metadata = {
        alias: _.get(metadata, 'cyclone.dev/alias', ''),
        description: _.get(metadata, 'cyclone.dev/description', ''),
      };
      defaultValue = _.merge(defaultValue, values);
    }

    return defaultValue;
  };

  render() {
    const { history } = this.props;
    if (this.props.project.detailLoading) {
      return <Spin />;
    }
    const initValue = this.getInitialValues();
    return (
      <Formik
        initialValues={initValue}
        validate={this.validate}
        onSubmit={this.submit}
        render={props => (
          <FormContent
            {...props}
            history={history}
            update={this.state.update}
          />
        )}
      />
    );
  }
}

AddProject.propTypes = {
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  project: PropTypes.object,
  history: PropTypes.object,
  match: PropTypes.object,
};

export default AddProject;

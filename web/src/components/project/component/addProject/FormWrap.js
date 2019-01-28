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
    const update = !!_.get(params, 'projectId');
    this.state = {
      update,
    };
    if (update) {
      props.project.getProject(params.projectId);
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
    const data = _.omit(values, 'metadata.description');
    data.spec.integrations = _.map(values.spec.integrations, n => {
      const resources = n.split('/');
      return { type: resources[0], name: resources[1] };
    });
    data.metadata.labels = {
      'cyclone.io/description': values.metadata.description,
    };
    if (update) {
      project.updateProject(data, params.projectId, () => {
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

    if (!values.metadata.name) {
      errors.metadata = { name: 'Required' };
    }
    return errors;
  };

  getInitialValues = () => {
    const { update } = this.state;
    let defaultValue = {
      metadata: { name: '', description: '' },
      spec: {
        integrations: [],
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
      const values = _.pick(proejctInfo, [
        'metadata.name',
        'spec.integrations',
        'spec.quota',
      ]);
      const description = _.get(proejctInfo, [
        'metadata',
        'labels',
        'cyclone.io/description',
      ]);
      values.spec.integrations = _.map(
        _.get(values, 'spec.integrations', []),
        n => `${n.type}/${n.name}`
      );
      values.metadata.description = description;
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

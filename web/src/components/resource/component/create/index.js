import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { toJS } from 'mobx';
import { inject, observer } from 'mobx-react';
import { Spin } from 'antd';
import FormContent from './FormContent';

@inject('resource')
@observer
class CreateOrUpdate extends React.Component {
  constructor(props) {
    super(props);
    const {
      match: { params },
    } = props;
    const update = !!_.get(params, 'resourceTypeName');
    this.state = {
      update,
    };
    if (update) {
      props.resource.getResourceType(params.resourceTypeName);
    }
  }

  getInitialValues = () => {
    const { update } = this.state;
    if (update) {
      return toJS(this.props.resource.resourceTypeDetail);
    }

    return { spec: { parameters: [] } };
  };

  validate = values => {
    const errors = {};
    if (!values.spec.type) {
      errors.spec = _.merge(errors.spec, {
        type: intl.get('resource.messages.typeRequired'),
      });
    }

    if (!values.spec.resolver) {
      errors.spec = _.merge(errors.spec, {
        resolver: intl.get('resource.messages.resolverRequired'),
      });
    }

    if (values.spec.parameters) {
      _.forEach(values.spec.parameters, value => {
        if (!value.name) {
          errors.spec = _.merge(errors.spec, {
            parameters: intl.get('resource.messages.parameterNameRequired'),
          });
        }
      });
    }

    return errors;
  };

  submit = values => {
    values.metadata = {
      name: `resource-type-${_.toLower(values.spec.type)}`,
      labels: {
        'cyclone.dev/builtin': 'true',
        'resource.cyclone.dev/template': 'true',
      },
    };

    const { update } = this.state;
    const { resource, history } = this.props;
    if (update) {
      resource.updateResourceType(values.spec.type, values, () => {
        history.replace(`/resource/${values.spec.type}`);
      });
    } else {
      resource.createResourceType(values, () => {
        history.replace('/resource');
      });
    }
  };

  render() {
    const { history, resource } = this.props;
    if (resource.resourceTypeLoading) {
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

CreateOrUpdate.propTypes = {
  resource: PropTypes.object,
  history: PropTypes.object,
  match: PropTypes.object,
  setFieldValue: PropTypes.func,
};

export default CreateOrUpdate;

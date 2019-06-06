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
    const { resource } = this.props;
    if (update) {
      const resourceTypeDetail = toJS(
        _.get(resource, 'resourceTypeDetail'),
        {}
      );
      if (resourceTypeDetail) {
        const paramBindings = _.get(
          resourceTypeDetail,
          'spec.bind.paramBindings',
          {}
        );
        const parameters = _.get(resourceTypeDetail, 'spec.parameters', []);
        parameters.forEach(v => {
          if (paramBindings[`${v.name}`]) {
            v.binding = paramBindings[`${v.name}`];
          }
        });
      }
      return resourceTypeDetail;
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
        'resource.cyclone.dev/template': 'true',
      },
    };

    const bindType = _.get(values, 'spec.bind.integrationType');
    const parameters = _.get(values, 'spec.parameters');

    if (bindType && parameters) {
      const paramBindings = {};
      parameters.forEach(v => {
        if (v.binding) {
          paramBindings[`${v.name}`] = v.binding;
        }
      });
      values.spec.bind = _.assign(values.spec.bind, paramBindings);
    }

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

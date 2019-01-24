import React from 'react';
import { Form, Input, Button } from 'antd';
import PropTypes from 'prop-types';
import { Field, Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import MakeField from '@/components/public/makeField';
import IntegrationSelect from './IntegrationSelect';
import ResourceAllocation from '@/components/public/resourceAllocation';

const { TextArea } = Input;

const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

@inject('project')
@observer
class AddProject extends React.Component {
  render() {
    const { project, history } = this.props;
    return (
      <Formik
        initialValues={{
          metadata: { name: '', description: '' },
          spec: {
            services: [],
            quota: {
              limits: {
                cpu: '',
                memory: '',
              },
              requests: {
                cpu: '',
                memory: '',
              },
            },
          },
        }}
        validate={values => {
          const errors = {};

          if (!values.metadata.name) {
            errors.metadata = { name: 'Required' };
          }
          return errors;
        }}
        onSubmit={values => {
          const data = { ...values };
          const services = _.map(values.spec.services, n => {
            const resources = n.split('/');
            return { type: resources[0], name: resources[1] };
          });
          data.spec.services = services;
          project.createProject(data, () => {
            history.replace(`/project`);
          });
        }}
        render={props => (
          <Form layout={'horizontal'} onSubmit={props.handleSubmit}>
            <Field
              label={intl.get('name')}
              name="metadata.name"
              component={InputField}
              hasFeedback
              required
            />
            <Field
              label={intl.get('description')}
              name="metadata.description"
              component={TextareaField}
            />
            <Field
              label={intl.get('project.externalSystem')}
              name="spec.services"
              required
              component={IntegrationSelect}
            />
            <Field
              label={intl.get('allocation.quotaConfig')}
              name="spec.quota"
              component={ResourceAllocation}
              onChange={value => {
                props.setFieldValue('spec.quota', value);
              }}
            />
            <div className="form-bottom-btn">
              <Button type="primary" htmlType="submit">
                {intl.get('operation.confirm')}
              </Button>
              <Button
                onClick={() => {
                  history.push(`/project`);
                }}
              >
                {intl.get('operation.cancel')}
              </Button>
            </div>
          </Form>
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
};

export default AddProject;

import React from 'react';
import { Form, Input, Button } from 'antd';
import PropTypes from 'prop-types';
import { Field, withFormik } from 'formik';
import MakeField from '@/public/makeField';
import IntegrationSelect from './IntegrationSelect';
import ResourceAllocation from '@/public/resourceAllocation';

const { TextArea } = Input;

const _Input = MakeField(Input);
const _Textarea = MakeField(TextArea);

const AddProject = props => {
  const { handleSubmit, setFieldValue } = props;
  const changeConfig = value => {
    setFieldValue('quota', value);
  };
  return (
    <Form onSubmit={handleSubmit} layout={'horizontal'}>
      <Field
        label={intl.get('name')}
        name="name"
        component={_Input}
        hasFeedback
        required
      />
      <Field
        label={intl.get('description')}
        name="description"
        component={_Textarea}
      />
      <Field
        label={'外部系统'}
        name="integration"
        required
        component={IntegrationSelect}
      />
      <Field
        label={intl.get('allocation.quotaConfig')}
        name="quota"
        component={ResourceAllocation}
        onChange={changeConfig}
      />
      <div className="form-bottom-btn">
        <Button type="primary" htmlType="submit">
          {intl.get('operation.confirm')}
        </Button>
        <Button>{intl.get('operation.cancel')}</Button>
      </div>
    </Form>
  );
};

AddProject.propTypes = {
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
};

export default withFormik({
  validate: values => {
    const errors = {};

    if (!values.name) {
      errors.name = 'Required';
    }

    return errors;
  },
  handleSubmit: values => {
    // TODO(qme): handle submit
  },
  displayName: 'addProject', // a unique identifier for this form
})(AddProject);

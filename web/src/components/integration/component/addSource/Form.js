import React from 'react';
import PropTypes from 'prop-types';
import { Form, Select, Input, Button } from 'antd';
import { withFormik, Field } from 'formik';
import ScmGroup from '../formGroup/ScmGroup';
import SonarQube from '../formGroup/SonarQube';
import DockerRegistry from '../formGroup/DockerRegistry';
import MakeField from '@/components/public/makeField';

const FormItem = Form.Item;
const Option = Select.Option;
const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const formMap = {
  SCM: <ScmGroup />,
  dockerregistry: <DockerRegistry />,
  sonarqube: <SonarQube />,
};

const selectSourceType = props => {
  return (
    <Select
      placeholder={intl.get('integration.form.validate.datasourcetype')}
      onChange={props.handleSelectChange}
    >
      <Option value="SCM">SCM</Option>
      <Option value="dockerregistry">
        {intl.get('integration.form.dockerregistry')}
      </Option>
      <Option value="sonarqube">SonarQube</Option>
    </Select>
  );
};
selectSourceType.propTypes = {
  handleSelectChange: PropTypes.func,
};

const SelectField = MakeField(selectSourceType);

class IntegrationForm extends React.Component {
  static propTypes = {
    form: PropTypes.object,
    payload: PropTypes.object,
    onSubmit: PropTypes.func,
    onCancel: PropTypes.func,
  };
  state = {
    formType: '',
  };

  handleSelectChange = value => {
    this.setState({
      formType: value,
    });
  };

  renderWrapForm = () => {
    const { formType } = this.state;
    return formMap[formType];
  };

  handleSelectChange = val => {
    this.setState({
      formType: val,
    });
  };

  render() {
    const { formType } = this.state;
    return (
      <Form>
        <Field
          label={intl.get('integration.name')}
          name="name"
          component={InputField}
          hasFeedback
          required
        />
        <Field
          label={intl.get('integration.desc')}
          name="desc"
          component={TextareaField}
        />
        <Field
          label={intl.get('integration.type')}
          name="type"
          required
          handleSelectChange={this.handleSelectChange}
          component={SelectField}
        />
        {formType && this.renderWrapForm()}
        <FormItem>
          <Button style={{ marginRight: '10px' }} onClick={this.handleCancle}>
            {intl.get('integration.form.cancel')}
          </Button>
          <Button type="primary" htmlType="submit">
            {intl.get('integration.form.confirm')}
          </Button>
        </FormItem>
      </Form>
    );
  }
}

export default withFormik({
  validate: values => {
    const errors = {};

    if (!values.name) {
      errors.name = 'Required';
    }

    return errors;
  },
  mapPropsToValues: () => ({ name: '', desc: '', type: '' }),
  handleSubmit: values => {},
})(IntegrationForm);

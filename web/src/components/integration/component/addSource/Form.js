import React from 'react';
import PropTypes from 'prop-types';
import { Form, Select, Input, Button } from 'antd';
import { withFormik, Field } from 'formik';
import ScmGroup from '../formGroup/ScmGroup/index';
import SonarQube from '../formGroup/SonarQube/index';
import DockerRegistry from '../formGroup/DockerRegistry/index';
import MakeField from '@/components/public/makeField';

const FormItem = Form.Item;
const Option = Select.Option;
const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

const selectSourceType = props => {
  return (
    <Select
      placeholder={intl.get('integration.form.datasourcetype')}
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
    payload: PropTypes.object,
    handleSubmit: PropTypes.func,
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
    const formMap = {
      SCM: <ScmGroup {...this.props} />,
      dockerregistry: <DockerRegistry />,
      sonarqube: <SonarQube />,
    };
    return formMap[formType];
  };

  handleSelectChange = val => {
    this.setState({
      formType: val,
    });
  };

  render() {
    const { formType } = this.state;
    const { handleSubmit } = this.props;
    return (
      <Form onSubmit={handleSubmit}>
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
          name="sourceType"
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
  mapPropsToValues: () => ({
    name: '',
    desc: '',
    sourceType: '',
    codeType: 'GitHub',
  }),
  handleSubmit: values => {
    console.log(values, 'elliot'); // eslint-disable-line
  },
})(IntegrationForm);

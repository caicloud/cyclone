import { Field } from 'formik';
import PropTypes from 'prop-types';
import { Form, Input, Button } from 'antd';
import MakeField from '@/components/public/makeField';
import ScmGroup from '../formGroup/ScmGroup';
import SonarQube from '../formGroup/SonarQube';
import DockerRegistry from '../formGroup/DockerRegistry';
import SelectSourceType from './SelectSourceType';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const renderWrapForm = (type, props) => {
  const formMap = {
    scm: <ScmGroup {...props} />,
    dockerRegistry: <DockerRegistry {...props} />,
    sonarQube: <SonarQube {...props} />,
  };
  return formMap[type];
};
const SelectField = MakeField(SelectSourceType);
const FormItem = Form.Item;
const FormContent = props => {
  const {
    values: {
      spec: { type },
    },
    handleSubmit,
    setFieldValue,
    handleCancle,
    errors,
  } = props;
  return (
    <Form onSubmit={handleSubmit}>
      <Field
        label={intl.get('integration.name')}
        name="metadata.alias"
        component={InputField}
        required
      />
      <Field
        label={intl.get('integration.desc')}
        name="metadata.description"
        component={TextareaField}
      />
      <Field
        label={intl.get('type')}
        name="spec.type"
        required
        handleSelectChange={val => {
          setFieldValue('spec.type', val);
        }}
        component={SelectField}
      />
      {type && renderWrapForm(type, { ...props })}
      <FormItem
        {...{
          labelCol: { span: 8 },
          wrapperCol: { span: 20 },
        }}
      >
        {!_.isEmpty(errors) && (
          <span className="tip-error">
            {intl.get('integration.form.tip.error')}
          </span>
        )}
        <Button style={{ float: 'right' }} type="primary" htmlType="submit">
          {intl.get('integration.form.confirm')}
        </Button>
        <Button
          style={{ float: 'right', marginRight: 10 }}
          onClick={handleCancle}
        >
          {intl.get('integration.form.cancel')}
        </Button>
      </FormItem>
    </Form>
  );
};

FormContent.propTypes = {
  history: PropTypes.object,
  errors: PropTypes.object,
  values: PropTypes.object,
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  handleCancle: PropTypes.func,
  update: PropTypes.bool,
};

export default FormContent;

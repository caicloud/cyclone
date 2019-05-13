import PropTypes from 'prop-types';
import { Form, Input, Button } from 'antd';
import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import InputSection from './Input';
import OutputSection from './Output';
import ConfigSection from './Config';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const FormItem = Form.Item;
const FormContent = props => {
  const { handleCancle, handleSubmit, update, values } = props;
  return (
    <Form onSubmit={handleSubmit}>
      <Field
        label={intl.get('name')}
        name="metadata.alias"
        component={InputField}
        disabled={update}
        required
      />
      <Field
        label={intl.get('description')}
        name="metadata.description"
        component={TextareaField}
      />
      <InputSection {...props} />
      <ConfigSection values={values} />
      <OutputSection {...props} />
      <FormItem
        {...{
          labelCol: { span: 8 },
          wrapperCol: { span: 20 },
        }}
      >
        <Button style={{ float: 'right' }} htmlType="submit" type="primary">
          {intl.get('confirm')}
        </Button>
        <Button
          style={{ float: 'right', marginRight: 10 }}
          onClick={handleCancle}
        >
          {intl.get('cancel')}
        </Button>
      </FormItem>
    </Form>
  );
};

FormContent.propTypes = {
  history: PropTypes.object,
  values: PropTypes.object,
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  handleCancle: PropTypes.func,
  update: PropTypes.bool,
};

export default FormContent;

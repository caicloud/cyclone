import { Input, Form } from 'antd';
import MakeField from '@/components/public/makeField';
import { Field } from 'formik';
import { required } from '@/components/public/validate';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class BasicInfo extends React.Component {
  render() {
    return (
      <Form>
        <Field
          label={intl.get('name')}
          name="metadata.name"
          component={InputField}
          hasFeedback
          required
          validate={required}
        />
        <Field
          label={intl.get('description')}
          name="metadata.annotations.description"
          component={TextareaField}
        />
      </Form>
    );
  }
}

export default BasicInfo;

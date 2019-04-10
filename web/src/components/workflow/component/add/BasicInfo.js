import { Input, Form } from 'antd';
import MakeField from '@/components/public/makeField';
import { Field } from 'formik';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class BasicInfo extends React.Component {
  render() {
    return (
      <Form>
        <Field
          label={intl.get('name')}
          name="metadata.alias"
          component={InputField}
          hasFeedback
          required
        />
        <Field
          label={intl.get('description')}
          name="metadata.description"
          component={TextareaField}
        />
      </Form>
    );
  }
}

export default BasicInfo;

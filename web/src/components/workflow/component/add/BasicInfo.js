import { Input, Form } from 'antd';
import { Field } from 'formik';
import PropTypes from 'prop-types';
import { defaultFormItemLayout } from '@/lib/const';
import { required } from '@/components/public/validate';
import MakeField from '@/components/public/makeField';

const FormItem = Form.Item;
const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class BasicInfo extends React.Component {
  static propTypes = {
    update: PropTypes.bool,
    values: PropTypes.object,
  };
  render() {
    const { update, values } = this.props;
    return (
      <Form>
        {update ? (
          <FormItem label={intl.get('name')} {...defaultFormItemLayout}>
            {_.get(values, `metadata.name`)}
          </FormItem>
        ) : (
          <Field
            label={intl.get('name')}
            name="metadata.name"
            component={update ? _.get(values, `metadata.name`) : InputField}
            hasFeedback
            required
            validate={required}
          />
        )}
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

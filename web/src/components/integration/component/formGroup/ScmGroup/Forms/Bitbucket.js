import { Field } from 'formik';
import { Input } from 'antd';
import MakeField from '@/components/public/makeField';
import ValidateSelect from './ValidateSelect';

const InputField = MakeField(Input);

const Bitbucket = props => {
  return (
    <div>
      <Field
        label={intl.get('integration.form.scm.serverAddress')}
        name="spec.scm.server"
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.scm.verificationMode')}
        {...props}
        name="spec.scm.validateType"
        component={ValidateSelect}
      />
    </div>
  );
};

export default Bitbucket;

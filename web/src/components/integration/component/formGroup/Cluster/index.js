import { Field } from 'formik';
import { Input } from 'antd';
import MakeField from '@/components/public/makeField';
import ValidateSelect from './ValidateSelect';

const InputField = MakeField(Input);

const Cluster = props => {
  return (
    <div className="u-cluster">
      <Field
        label={intl.get('integration.form.cluster.serverAddress')}
        name="spec.cluster.credential.server"
        required
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.cluster.verificationMode')}
        {...props}
        name="spec.cluster.credential.validateType"
        component={ValidateSelect}
      />
    </div>
  );
};

export default Cluster;

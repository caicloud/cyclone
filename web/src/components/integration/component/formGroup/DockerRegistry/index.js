import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { Input } from 'antd';

const InputField = MakeField(Input);

const DockerRegistry = () => {
  return (
    <div>
      <Field
        label={intl.get('integration.form.dockerRegistry.registryAddress')}
        required
        name="registryAddress"
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.dockerRegistry.admin')}
        required
        name="adminUser"
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.dockerRegistry.adminpwd')}
        required
        name="adminPwd"
        component={InputField}
      />
    </div>
  );
};

export default DockerRegistry;

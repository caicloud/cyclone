import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import PropTypes from 'prop-types';
import { Form, Input } from 'antd';
import SwitchField from './SwitchField';
const InputField = MakeField(Input);
const FormItem = Form.Item;

export default class IsWorkerCluster extends React.Component {
  static propTypes = {
    values: PropTypes.object,
    field: PropTypes.object,
    setFieldValue: PropTypes.func,
  };
  handleType = e => {
    const {
      setFieldValue,
      field: { name },
    } = this.props;
    const value = e.target.value;
    setFieldValue(name, value);
  };
  render() {
    const {
      values: {
        spec: {
          cluster: { isWorkerCluster },
        },
      },
      setFieldValue,
    } = this.props;
    return (
      <div>
        <Field
          label={intl.get('integration.form.cluster.isWorkerCluster')}
          {...this.props}
          onChange={val => {
            setFieldValue('spec.cluster.isWorkerCluster', val);
          }}
          name="spec.cluster.isWorkerCluster"
          component={SwitchField}
        />
        {isWorkerCluster && (
          <FormItem>
            <Field
              label={intl.get('integration.form.cluster.namespace')}
              name="spec.cluster.namespace"
              component={InputField}
            />
            <Field
              label={intl.get('integration.form.cluster.pvc')}
              name="spec.cluster.pvc"
              component={InputField}
            />
          </FormItem>
        )}
      </div>
    );
  }
}

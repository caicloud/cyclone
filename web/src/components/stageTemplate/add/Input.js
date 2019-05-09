import PropTypes from 'prop-types';
import SectionCard from '@/components/public/sectionCard';
import MakeField from '@/components/public/makeField';
import SelectSourceType from './SelectSourceType';
import { Field } from 'formik';
import { Form } from 'antd';

const SelectField = MakeField(SelectSourceType);
const FormItem = Form.Item;

const InputSection = props => {
  const { setFieldValue } = props;
  return (
    <SectionCard title={intl.get('input')}>
      <FormItem>
        <Field
          label={intl.get('type')}
          name="spec.pod.inputs.resources"
          required
          handleSelectChange={val => {
            setFieldValue('spec.pod.inputs.resources', [
              {
                name: '',
                type: val,
                path: '',
              },
            ]);
          }}
          component={SelectField}
        />
      </FormItem>
    </SectionCard>
  );
};

InputSection.propTypes = {
  setFieldValue: PropTypes.func,
};

export default InputSection;

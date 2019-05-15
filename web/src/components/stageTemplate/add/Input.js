import PropTypes from 'prop-types';
import SectionCard from '@/components/public/sectionCard';
import ResourceType from './ResourceType';

const InputSection = props => {
  const { values, setFieldValue, errors } = props;
  return (
    <SectionCard title={intl.get('input')}>
      <ResourceType
        required
        path="spec.pod.inputs.resources"
        values={values}
        setFieldValue={setFieldValue}
        errors={errors}
      />
    </SectionCard>
  );
};

InputSection.propTypes = {
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
  errors: PropTypes.object,
};

export default InputSection;

import PropTypes from 'prop-types';
import SectionCard from '@/components/public/sectionCard';
import ResourceType from './ResourceType';

const InputSection = props => {
  const { values, setFieldValue } = props;
  return (
    <SectionCard title={intl.get('input')}>
      <ResourceType
        required
        path="spec.pod.inputs.resources"
        values={values}
        setFieldValue={setFieldValue}
      />
    </SectionCard>
  );
};

InputSection.propTypes = {
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
};

export default InputSection;

import { Form, Input } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { noLabelItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import SelectPlus from '@/components/public/makeField/select';

const FormItem = Form.Item;
const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);

class SCM extends React.Component {
  render() {
    const { values, integrationList, setFieldValue } = this.props;
    return (
      <FormItem {...noLabelItemLayout}>
        <FieldArray
          name="spec.parameters"
          render={() => (
            <div>
              {_.get(values, 'spec.parameters', []).map((field, index) => {
                if (field.name === 'SCM_TOKEN') {
                  return (
                    <Field
                      key={field.name}
                      label={'集成中心'}
                      name={`spec.parameters.${index}.value`}
                      payload={{
                        items: integrationList,
                        nameKey: 'metadata.name',
                        valueKey: 'spec.scm.token',
                      }}
                      handleSelectChange={val => {
                        setFieldValue(`spec.parameters.${index}.value`, val);
                      }}
                      component={SelectField}
                    />
                  );
                } else {
                  return (
                    <Field
                      key={field.name}
                      label={
                        field.name.includes('SCM_')
                          ? field.name.replace('SCM_', '')
                          : field.name
                      }
                      name={`spec.parameters.${index}.value`}
                      component={InputField}
                      hasFeedback
                      required
                    />
                  );
                }
              })}
            </div>
          )}
        />
      </FormItem>
    );
  }
}

SCM.propTypes = {
  values: PropTypes.object,
  integrationList: PropTypes.array,
  setFieldValue: PropTypes.func,
};

export default SCM;

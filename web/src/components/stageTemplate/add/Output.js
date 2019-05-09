import PropTypes from 'prop-types';
import SectionCard from '@/components/public/sectionCard';
import MakeField from '@/components/public/makeField';
import { defaultFormItemLayout } from '@/lib/const';
import { Field, FieldArray } from 'formik';
import { Form, Input, Row, Col, Button } from 'antd';
import SelectSourceType from './SelectSourceType';

const InputField = MakeField(Input);
const FormItem = Form.Item;
const SelectField = MakeField(SelectSourceType);

const OutputSection = props => {
  const { setFieldValue, values } = props;
  return (
    <SectionCard title={intl.get('output')}>
      <Field
        label={intl.get('type')}
        name="spec.pod.outputs.resources"
        required
        handleSelectChange={val => {
          setFieldValue('spec.pod.outputs.resources', [
            {
              name: '',
              type: val,
              path: '',
            },
          ]);
        }}
        component={SelectField}
      />
      <FormItem label={'Artifact'} {...defaultFormItemLayout}>
        <FieldArray
          name={'spec.pod.outputs.artifacts'}
          render={arrayHelpers => (
            <div>
              {_.get(values, 'spec.pod.outputs.artifacts', []).length > 0 && (
                <Row gutter={16}>
                  <Col span={11}>{intl.get('name')}</Col>
                  <Col span={11}>{intl.get('path')}</Col>
                </Row>
              )}
              {_.get(values, 'spec.pod.outputs.artifacts', []).map(
                (a, index) => (
                  <Row key={index} gutter={16}>
                    <Col span={11}>
                      <Field
                        key={a.name}
                        name={`spec.pod.outputs.artifacts.${index}.name`}
                        component={InputField}
                        hasFeedback
                      />
                    </Col>
                    <Col span={11}>
                      <Field
                        key={a.value}
                        name={`spec.pod.outputs.artifacts.${index}.path`}
                        component={InputField}
                        hasFeedback
                      />
                    </Col>
                    <Col span={2}>
                      <Button
                        type="circle"
                        icon="delete"
                        onClick={() => arrayHelpers.remove(index)}
                      />
                    </Col>
                  </Row>
                )
              )}
              <Button
                ico="plus"
                onClick={() => arrayHelpers.push({ name: '', path: '' })}
              >
                {intl.get('workflow.addArtifact')}
              </Button>
            </div>
          )}
        />
      </FormItem>
    </SectionCard>
  );
};

OutputSection.propTypes = {
  values: PropTypes.object,
  setFieldValue: PropTypes.func,
};

export default OutputSection;

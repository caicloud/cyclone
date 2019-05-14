import PropTypes from 'prop-types';
import MakeField from '@/components/public/makeField';
import SelectSourceType from './SelectSourceType';
import { defaultFormItemLayout } from '@/lib/const';
import { Field, FieldArray } from 'formik';
import { Form, Row, Col, Button } from 'antd';
const FormItem = Form.Item;

const SelectField = MakeField(SelectSourceType);

// TODO huxt validate resource type select
const ResourceType = props => {
  const { setFieldValue, values, path, required } = props;
  const resources = _.get(values, path, []);
  return (
    <FieldArray
      name={path}
      render={arrayHelpers => (
        <FormItem
          required={required}
          label={intl.get('template.resourceType')}
          {...defaultFormItemLayout}
        >
          {resources.map((a, index) => (
            <Row key={index} gutter={10}>
              <Col span={22}>
                <Field
                  key={a}
                  name={`${path}.${index}`}
                  handleSelectChange={val => {
                    setFieldValue(`${path}.${index}`, {
                      name: '',
                      type: val,
                      path: '',
                    });
                  }}
                  component={SelectField}
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
          ))}
          <Button
            ico="plus"
            onClick={() => {
              arrayHelpers.push({ name: '', type: 'Git', path: '' });
            }}
          >
            {intl.get('template.addResource')}
          </Button>
        </FormItem>
      )}
    />
  );
};

ResourceType.propTypes = {
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
  path: PropTypes.string,
  touched: PropTypes.object,
  errors: PropTypes.object,
  required: PropTypes.bool,
};

export default ResourceType;

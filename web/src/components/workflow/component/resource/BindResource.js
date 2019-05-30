import { Form, Input } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { noLabelItemLayout, modalFormItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { required } from '@/components/public/validate';
import SelectPlus from '@/components/public/makeField/select';

const FormItem = Form.Item;
const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);

const Fragment = React.Fragment;

// use in add stage, select a exist resource or create a new resource
@inject('integration', 'resource')
@observer
class BindResource extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    type: PropTypes.oneOf(['inputs', 'outputs']),
    integration: PropTypes.object,
    update: PropTypes.bool,
    resource: PropTypes.object,
  };

  state = { addWay: 'new' };

  componentDidMount() {
    // TODO(qme): user integration
    // const { integration } = this.props;
    // integration.getIntegrationList();
  }

  changeAddWay = value => {
    this.setState({ addWay: value });
  };

  handleTypeChange = val => {
    const {
      setFieldValue,
      resource: { resourceTypeList },
    } = this.props;
    setFieldValue('spec.type', val);
    const item = _.find(
      _.get(resourceTypeList, 'items', []),
      o => _.get(o, 'spec.type') === val
    );
    if (item) {
      setFieldValue('spec.parameters', _.get(item, 'spec.parameters'));
    }
  };

  render() {
    const { addWay } = this.state;
    const {
      values,
      setFieldValue,
      type,
      update,
      resource: { resourceTypeList },
    } = this.props;
    return (
      <Form layout={'horizontal'}>
        <Field
          label={intl.get('workflow.resourceType')}
          name="spec.type"
          handleSelectChange={val => {
            this.handleTypeChange(val);
          }}
          payload={{
            items: _.get(resourceTypeList, 'items', []),
            nameKey: 'spec.type',
            valueKey: 'spec.type',
          }}
          component={SelectField}
          formItemLayout={modalFormItemLayout}
          required
          validate={required}
        />
        {addWay === 'exist' ? (
          <Field
            label={intl.get('type')}
            name="name"
            handleSelectChange={val => {
              setFieldValue('name', val);
            }}
            component={<div>TODO: resource select</div>}
            formItemLayout={modalFormItemLayout}
            required
            validate={required}
          />
        ) : (
          <Fragment>
            {update ? (
              <FormItem label={intl.get('name')} {...modalFormItemLayout}>
                {_.get(values, 'name')}
              </FormItem>
            ) : (
              <Field
                label={intl.get('name')}
                name="name"
                component={InputField}
                formItemLayout={modalFormItemLayout}
                hasFeedback
                required
                validate={required}
              />
            )}
            <FormItem {...noLabelItemLayout}>
              <FieldArray
                name="spec.parameters"
                render={() => (
                  <div>
                    {_.get(values, 'spec.parameters', []).map(
                      (field, index) => (
                        <Field
                          key={field.name}
                          label={
                            field.name.includes('SCM_')
                              ? field.name.replace('SCM_', '')
                              : field.name
                          }
                          name={`spec.parameters.${index}.value`}
                          component={InputField}
                          formItemLayout={modalFormItemLayout}
                          hasFeedback
                          required
                          tooltip={field.description}
                          validate={required}
                        />
                      )
                    )}
                  </div>
                )}
              />
            </FormItem>
          </Fragment>
        )}
        {type === 'inputs' && (
          <Field
            label={intl.get('workflow.usePath')}
            name="path"
            component={InputField}
            hasFeedback
            required
            validate={required}
            formItemLayout={modalFormItemLayout}
          />
        )}
      </Form>
    );
  }
}

export default BindResource;

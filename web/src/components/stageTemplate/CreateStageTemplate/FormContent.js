import { Field } from 'formik';
import KeyValue from '@/components/public/KeyValue';
import PropTypes from 'prop-types';
import { Form, Input, Button } from 'antd';
import MakeField from '@/components/public/makeField';
import SelectSourceType from './SelectSourceType';
import style from './template.module.less';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const SelectField = MakeField(SelectSourceType);
const FormItem = Form.Item;
const FormContent = props => {
  const { setFieldValue, handleCancle, submit, update } = props;
  return (
    <Form>
      <Field
        label={intl.get('integration.name')}
        name="metadata.alias"
        component={InputField}
        disabled={update}
        required
      />
      <Field
        label={intl.get('integration.desc')}
        name="metadata.description"
        component={TextareaField}
      />
      <h3>输入</h3>
      <Field
        label={intl.get('type')}
        name="spec.type"
        required
        handleSelectChange={val => {
          setFieldValue('spec.type', val);
        }}
        component={SelectField}
      />
      <h3>{intl.get('template.form.config.name')}</h3>
      <Field
        label={intl.get('template.form.config.registory')}
        name="metadata.alias"
        component={InputField}
        required
      />
      <Field
        label={intl.get('template.form.config.entryPoint')}
        name="metadata.description"
        component={InputField}
      />
      <Field
        label={intl.get('template.form.config.CMD')}
        name="metadata.description"
        component={InputField}
      />
      <KeyValue
        cls={style['kv-item']}
        name={intl.get('stage.spec.container.image')}
        value={'asdadsasda'}
      />
      <FormItem>
        <div>sad</div>
        <div>sadsddd</div>
      </FormItem>
      <h3>输出</h3>
      <Field
        label={intl.get('type')}
        name="spec.type"
        required
        handleSelectChange={val => {
          setFieldValue('spec.type', val);
        }}
        component={SelectField}
      />
      <FormItem
        {...{
          labelCol: { span: 8 },
          wrapperCol: { span: 20 },
        }}
      >
        <Button style={{ float: 'right' }} onClick={submit} type="primary">
          {intl.get('integration.form.confirm')}
        </Button>
        <Button
          style={{ float: 'right', marginRight: 10 }}
          onClick={handleCancle}
        >
          {intl.get('integration.form.cancel')}
        </Button>
      </FormItem>
    </Form>
  );
};

FormContent.propTypes = {
  history: PropTypes.object,
  values: PropTypes.object,
  submit: PropTypes.func,
  setFieldValue: PropTypes.func,
  handleCancle: PropTypes.func,
  update: PropTypes.bool,
};

export default FormContent;

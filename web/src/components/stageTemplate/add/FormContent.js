import PropTypes from 'prop-types';
import { Form, Input, Button } from 'antd';
import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { inject, observer } from 'mobx-react';
import AutoCompletePlus from '@/components/public/makeField/autoComplete';
import InputSection from './Input';
import OutputSection from './Output';
import ConfigSection from './Config';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const AutoCompleteField = MakeField(AutoCompletePlus);
const FormItem = Form.Item;
@inject('stageTemplate')
@observer
class FormContent extends React.Component {
  componentDidMount() {
    this.props.stageTemplate.getTemplateList();
  }
  getKinds = list => {
    const kinds = [];
    _.forEach(list, item => {
      const kind = _.get(item, [
        'metadata',
        'labels',
        'stage.cyclone.dev/template-kind',
      ]);
      kind && !kinds.includes(kind) && kinds.push(kind);
    });
    return _.map(kinds, kind => ({
      alias: intl.get(`template.kinds.${kind}`) || kind,
      value: kind,
    }));
  };
  render() {
    const {
      handleCancle,
      handleSubmit,
      update,
      values,
      setFieldValue,
      stageTemplate: { templateList = [] },
    } = this.props;
    const kinds = this.getKinds(templateList.items);

    return (
      <Form onSubmit={handleSubmit}>
        <Field
          label={intl.get('name')}
          name="metadata.name"
          component={InputField}
          disabled={update}
          placeholder={intl.get('template.form.placeholder.name')}
          required
        />
        <Field
          label={intl.get('template.form.scene')}
          name="metadata.scene"
          component={InputField}
          placeholder={intl.get('template.form.placeholder.scene')}
        />
        <Field
          label={intl.get('template.form.kind')}
          name="metadata.kind"
          payload={{
            items: kinds,
            nameKey: 'alias',
          }}
          handleSelectChange={val => {
            setFieldValue('metadata.kind', val);
          }}
          placeholder={intl.get('template.form.placeholder.kind')}
          component={AutoCompleteField}
        />
        <Field
          label={intl.get('description')}
          name="metadata.description"
          component={TextareaField}
          placeholder={intl.get('template.form.placeholder.desc')}
        />
        <InputSection {...this.props} />
        <ConfigSection values={values} />
        <OutputSection {...this.props} />
        <FormItem
          {...{
            labelCol: { span: 8 },
            wrapperCol: { span: 20 },
          }}
        >
          <Button style={{ float: 'right' }} htmlType="submit" type="primary">
            {intl.get('confirm')}
          </Button>
          <Button
            style={{ float: 'right', marginRight: 10 }}
            onClick={handleCancle}
          >
            {intl.get('cancel')}
          </Button>
        </FormItem>
      </Form>
    );
  }
}

FormContent.propTypes = {
  history: PropTypes.object,
  values: PropTypes.object,
  stageTemplate: PropTypes.object,
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  handleCancle: PropTypes.func,
  update: PropTypes.bool,
};

export default FormContent;

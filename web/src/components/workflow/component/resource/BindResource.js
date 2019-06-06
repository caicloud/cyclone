import { Form, Input } from 'antd';
import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { modalFormItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { required } from '@/components/public/validate';
import SelectPlus from '@/components/public/makeField/select';

const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);

// use in add stage, select a exist resource or create a new resource
@inject('resource', 'project')
@observer
class BindResource extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    type: PropTypes.oneOf(['inputs', 'outputs']),
    resource: PropTypes.object,
    projectName: PropTypes.string,
    project: PropTypes.object,
  };

  constructor(props) {
    super(props);
    const {
      project: { resourceList },
    } = props;
    this.state = {
      resourceMap: this.getResourceMap(_.get(resourceList, 'items', [])),
    };
  }

  componentDidMount() {
    const {
      project: { resourceList, listProjectResources },
      projectName,
    } = this.props;
    if (!resourceList) {
      listProjectResources(projectName, data => {
        const resourceMap = this.getResourceMap(_.get(data, 'items', []));
        this.setState({ resourceMap });
      });
    }
  }

  getResourceMap = list => {
    let obj = {};
    _.forEach(list, v => {
      const type = _.get(v, 'spec.type');
      if (!obj[type]) {
        obj[type] = [];
      }
      obj[type].push(v);
    });
    return obj;
  };

  render() {
    const {
      values,
      type,
      resource: { resourceTypeList },
      setFieldValue,
    } = this.props;
    const { resourceMap } = this.state;
    const resourceType = _.get(values, 'type');
    const list = _.get(resourceMap, resourceType, []);
    return (
      <Form layout={'horizontal'}>
        <Field
          label={intl.get('workflow.resourceType')}
          name="type"
          handleSelectChange={val => {
            setFieldValue('type', val);
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
        <Field
          label={intl.get('resource.selectResource')}
          name="name"
          handleSelectChange={val => {
            setFieldValue('name', val);
          }}
          component={SelectField}
          payload={{
            items: list,
            nameKey: 'metadata.name',
            valueKey: 'metadata.name',
          }}
          formItemLayout={modalFormItemLayout}
          required
          validate={required}
        />
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

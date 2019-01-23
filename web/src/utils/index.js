import { Form } from 'antd';

/**
 * 创建FormItem回显到表单的对象
 * @param obj
 * @returns {{}}
 */
export function createFormItemObj(obj) {
  let target = {};
  for (let [key, value] of Object.entries(obj)) {
    target[key] = Form.createFormField(value);
  }
  return target;
}

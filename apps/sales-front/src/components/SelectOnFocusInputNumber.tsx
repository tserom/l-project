import { InputNumber, type InputNumberProps } from 'antd'

/** antd InputNumber：聚焦全选，方便直接覆盖默认 0 / 0.00 */
export default function SelectOnFocusInputNumber(props: InputNumberProps) {
  const { onFocus, ...rest } = props
  return (
    <InputNumber
      {...rest}
      onFocus={(e) => {
        onFocus?.(e)
        const el = e.target
        requestAnimationFrame(() => el.select())
      }}
    />
  )
}

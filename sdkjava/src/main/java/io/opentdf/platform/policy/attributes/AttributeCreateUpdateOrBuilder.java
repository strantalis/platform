// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: policy/attributes/attributes.proto

// Protobuf Java Version: 3.25.3
package io.opentdf.platform.policy.attributes;

public interface AttributeCreateUpdateOrBuilder extends
    // @@protoc_insertion_point(interface_extends:policy.attributes.AttributeCreateUpdate)
    com.google.protobuf.MessageOrBuilder {

  /**
   * <pre>
   * Optional metadata for the attribute definition
   * </pre>
   *
   * <code>.common.MetadataMutable metadata = 1 [json_name = "metadata"];</code>
   * @return Whether the metadata field is set.
   */
  boolean hasMetadata();
  /**
   * <pre>
   * Optional metadata for the attribute definition
   * </pre>
   *
   * <code>.common.MetadataMutable metadata = 1 [json_name = "metadata"];</code>
   * @return The metadata.
   */
  io.opentdf.platform.common.MetadataMutable getMetadata();
  /**
   * <pre>
   * Optional metadata for the attribute definition
   * </pre>
   *
   * <code>.common.MetadataMutable metadata = 1 [json_name = "metadata"];</code>
   */
  io.opentdf.platform.common.MetadataMutableOrBuilder getMetadataOrBuilder();

  /**
   * <pre>
   * namespace of the attribute
   * </pre>
   *
   * <code>string namespace_id = 2 [json_name = "namespaceId", (.buf.validate.field) = { ... }</code>
   * @return The namespaceId.
   */
  java.lang.String getNamespaceId();
  /**
   * <pre>
   * namespace of the attribute
   * </pre>
   *
   * <code>string namespace_id = 2 [json_name = "namespaceId", (.buf.validate.field) = { ... }</code>
   * @return The bytes for namespaceId.
   */
  com.google.protobuf.ByteString
      getNamespaceIdBytes();

  /**
   * <pre>
   *attribute name
   * </pre>
   *
   * <code>string name = 3 [json_name = "name", (.buf.validate.field) = { ... }</code>
   * @return The name.
   */
  java.lang.String getName();
  /**
   * <pre>
   *attribute name
   * </pre>
   *
   * <code>string name = 3 [json_name = "name", (.buf.validate.field) = { ... }</code>
   * @return The bytes for name.
   */
  com.google.protobuf.ByteString
      getNameBytes();

  /**
   * <pre>
   * attribute rule enum
   * </pre>
   *
   * <code>.policy.attributes.AttributeRuleTypeEnum rule = 4 [json_name = "rule", (.buf.validate.field) = { ... }</code>
   * @return The enum numeric value on the wire for rule.
   */
  int getRuleValue();
  /**
   * <pre>
   * attribute rule enum
   * </pre>
   *
   * <code>.policy.attributes.AttributeRuleTypeEnum rule = 4 [json_name = "rule", (.buf.validate.field) = { ... }</code>
   * @return The rule.
   */
  io.opentdf.platform.policy.attributes.AttributeRuleTypeEnum getRule();

  /**
   * <pre>
   * optional
   * </pre>
   *
   * <code>repeated .policy.attributes.ValueCreateUpdate values = 5 [json_name = "values"];</code>
   */
  java.util.List<io.opentdf.platform.policy.attributes.ValueCreateUpdate> 
      getValuesList();
  /**
   * <pre>
   * optional
   * </pre>
   *
   * <code>repeated .policy.attributes.ValueCreateUpdate values = 5 [json_name = "values"];</code>
   */
  io.opentdf.platform.policy.attributes.ValueCreateUpdate getValues(int index);
  /**
   * <pre>
   * optional
   * </pre>
   *
   * <code>repeated .policy.attributes.ValueCreateUpdate values = 5 [json_name = "values"];</code>
   */
  int getValuesCount();
  /**
   * <pre>
   * optional
   * </pre>
   *
   * <code>repeated .policy.attributes.ValueCreateUpdate values = 5 [json_name = "values"];</code>
   */
  java.util.List<? extends io.opentdf.platform.policy.attributes.ValueCreateUpdateOrBuilder> 
      getValuesOrBuilderList();
  /**
   * <pre>
   * optional
   * </pre>
   *
   * <code>repeated .policy.attributes.ValueCreateUpdate values = 5 [json_name = "values"];</code>
   */
  io.opentdf.platform.policy.attributes.ValueCreateUpdateOrBuilder getValuesOrBuilder(
      int index);
}
// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: policy/subjectmapping/subject_mapping.proto

// Protobuf Java Version: 3.25.3
package io.opentdf.platform.policy.subjectmapping;

public final class SubjectMappingProto {
  private SubjectMappingProto() {}
  public static void registerAllExtensions(
      com.google.protobuf.ExtensionRegistryLite registry) {
  }

  public static void registerAllExtensions(
      com.google.protobuf.ExtensionRegistry registry) {
    registerAllExtensions(
        (com.google.protobuf.ExtensionRegistryLite) registry);
  }
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_Condition_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_Condition_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_ConditionGroup_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_ConditionGroup_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_SubjectSet_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_SubjectSet_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_SubjectMapping_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_SubjectMapping_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_Subject_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_Subject_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_MatchSubjectMappingsRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_MatchSubjectMappingsRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_MatchSubjectMappingsResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_MatchSubjectMappingsResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_SubjectMappingCreateUpdate_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_SubjectMappingCreateUpdate_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_GetSubjectMappingRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_GetSubjectMappingRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_GetSubjectMappingResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_GetSubjectMappingResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_ListSubjectMappingsRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_ListSubjectMappingsRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_ListSubjectMappingsResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_ListSubjectMappingsResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_CreateSubjectMappingRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_CreateSubjectMappingRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_CreateSubjectMappingResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_CreateSubjectMappingResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_UpdateSubjectMappingRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_UpdateSubjectMappingRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_UpdateSubjectMappingResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_UpdateSubjectMappingResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_DeleteSubjectMappingRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_DeleteSubjectMappingRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_policy_subjectmapping_DeleteSubjectMappingResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_policy_subjectmapping_DeleteSubjectMappingResponse_fieldAccessorTable;

  public static com.google.protobuf.Descriptors.FileDescriptor
      getDescriptor() {
    return descriptor;
  }
  private static  com.google.protobuf.Descriptors.FileDescriptor
      descriptor;
  static {
    java.lang.String[] descriptorData = {
      "\n+policy/subjectmapping/subject_mapping." +
      "proto\022\025policy.subjectmapping\032!authorizat" +
      "ion/authorization.proto\032\033buf/validate/va" +
      "lidate.proto\032\023common/common.proto\032\034googl" +
      "e/api/annotations.proto\032\034google/protobuf" +
      "/struct.proto\032\"policy/attributes/attribu" +
      "tes.proto\"\325\001\n\tCondition\0224\n\026subject_exter" +
      "nal_field\030\001 \001(\tR\024subjectExternalField\022Z\n" +
      "\010operator\030\002 \001(\01621.policy.subjectmapping." +
      "SubjectMappingOperatorEnumB\013\272H\010\202\001\002\020\001\310\001\001R" +
      "\010operator\0226\n\027subject_external_values\030\003 \003" +
      "(\tR\025subjectExternalValues\"\275\001\n\016ConditionG" +
      "roup\022J\n\nconditions\030\001 \003(\0132 .policy.subjec" +
      "tmapping.ConditionB\010\272H\005\222\001\002\010\001R\nconditions" +
      "\022_\n\014boolean_type\030\002 \001(\0162/.policy.subjectm" +
      "apping.ConditionBooleanTypeEnumB\013\272H\010\202\001\002\020" +
      "\001\310\001\001R\013booleanType\"h\n\nSubjectSet\022Z\n\020condi" +
      "tion_groups\030\001 \003(\0132%.policy.subjectmappin" +
      "g.ConditionGroupB\010\272H\005\222\001\002\010\001R\017conditionGro" +
      "ups\"\210\002\n\016SubjectMapping\022\016\n\002id\030\001 \001(\tR\002id\022," +
      "\n\010metadata\030\002 \001(\0132\020.common.MetadataR\010meta" +
      "data\022A\n\017attribute_value\030\003 \001(\0132\030.policy.a" +
      "ttributes.ValueR\016attributeValue\022D\n\014subje" +
      "ct_sets\030\004 \003(\0132!.policy.subjectmapping.Su" +
      "bjectSetR\013subjectSets\022/\n\007actions\030\005 \003(\0132\025" +
      ".authorization.ActionR\007actions\"B\n\007Subjec" +
      "t\0227\n\nattributes\030\001 \001(\0132\027.google.protobuf." +
      "StructR\nattributes\"W\n\033MatchSubjectMappin" +
      "gsRequest\0228\n\007subject\030\001 \001(\0132\036.policy.subj" +
      "ectmapping.SubjectR\007subject\"p\n\034MatchSubj" +
      "ectMappingsResponse\022P\n\020subject_mappings\030" +
      "\001 \003(\0132%.policy.subjectmapping.SubjectMap" +
      "pingR\017subjectMappings\"\366\001\n\032SubjectMapping" +
      "CreateUpdate\0223\n\010metadata\030\001 \001(\0132\027.common." +
      "MetadataMutableR\010metadata\022,\n\022attribute_v" +
      "alue_id\030\002 \001(\tR\020attributeValueId\022D\n\014subje" +
      "ct_sets\030\003 \003(\0132!.policy.subjectmapping.Su" +
      "bjectSetR\013subjectSets\022/\n\007actions\030\004 \003(\0132\025" +
      ".authorization.ActionR\007actions\"2\n\030GetSub" +
      "jectMappingRequest\022\026\n\002id\030\001 \001(\tB\006\272H\003\310\001\001R\002" +
      "id\"k\n\031GetSubjectMappingResponse\022N\n\017subje" +
      "ct_mapping\030\001 \001(\0132%.policy.subjectmapping" +
      ".SubjectMappingR\016subjectMapping\"\034\n\032ListS" +
      "ubjectMappingsRequest\"o\n\033ListSubjectMapp" +
      "ingsResponse\022P\n\020subject_mappings\030\001 \003(\0132%" +
      ".policy.subjectmapping.SubjectMappingR\017s" +
      "ubjectMappings\"\201\001\n\033CreateSubjectMappingR" +
      "equest\022b\n\017subject_mapping\030\001 \001(\01321.policy" +
      ".subjectmapping.SubjectMappingCreateUpda" +
      "teB\006\272H\003\310\001\001R\016subjectMapping\"n\n\034CreateSubj" +
      "ectMappingResponse\022N\n\017subject_mapping\030\001 " +
      "\001(\0132%.policy.subjectmapping.SubjectMappi" +
      "ngR\016subjectMapping\"\231\001\n\033UpdateSubjectMapp" +
      "ingRequest\022\026\n\002id\030\001 \001(\tB\006\272H\003\310\001\001R\002id\022b\n\017su" +
      "bject_mapping\030\002 \001(\01321.policy.subjectmapp" +
      "ing.SubjectMappingCreateUpdateB\006\272H\003\310\001\001R\016" +
      "subjectMapping\"n\n\034UpdateSubjectMappingRe" +
      "sponse\022N\n\017subject_mapping\030\001 \001(\0132%.policy" +
      ".subjectmapping.SubjectMappingR\016subjectM" +
      "apping\"5\n\033DeleteSubjectMappingRequest\022\026\n" +
      "\002id\030\001 \001(\tB\006\272H\003\310\001\001R\002id\"n\n\034DeleteSubjectMa" +
      "ppingResponse\022N\n\017subject_mapping\030\001 \001(\0132%" +
      ".policy.subjectmapping.SubjectMappingR\016s" +
      "ubjectMapping*\233\001\n\032SubjectMappingOperator" +
      "Enum\022-\n)SUBJECT_MAPPING_OPERATOR_ENUM_UN" +
      "SPECIFIED\020\000\022$\n SUBJECT_MAPPING_OPERATOR_" +
      "ENUM_IN\020\001\022(\n$SUBJECT_MAPPING_OPERATOR_EN" +
      "UM_NOT_IN\020\002*\220\001\n\030ConditionBooleanTypeEnum" +
      "\022+\n\'CONDITION_BOOLEAN_TYPE_ENUM_UNSPECIF" +
      "IED\020\000\022#\n\037CONDITION_BOOLEAN_TYPE_ENUM_AND" +
      "\020\001\022\"\n\036CONDITION_BOOLEAN_TYPE_ENUM_OR\020\0022\371" +
      "\007\n\025SubjectMappingService\022\251\001\n\024MatchSubjec" +
      "tMappings\0222.policy.subjectmapping.MatchS" +
      "ubjectMappingsRequest\0323.policy.subjectma" +
      "pping.MatchSubjectMappingsResponse\"(\202\323\344\223" +
      "\002\"\"\027/subject-mappings/match:\007subject\022\227\001\n" +
      "\023ListSubjectMappings\0221.policy.subjectmap" +
      "ping.ListSubjectMappingsRequest\0322.policy" +
      ".subjectmapping.ListSubjectMappingsRespo" +
      "nse\"\031\202\323\344\223\002\023\022\021/subject-mappings\022\226\001\n\021GetSu" +
      "bjectMapping\022/.policy.subjectmapping.Get" +
      "SubjectMappingRequest\0320.policy.subjectma" +
      "pping.GetSubjectMappingResponse\"\036\202\323\344\223\002\030\022" +
      "\026/subject-mappings/{id}\022\253\001\n\024CreateSubjec" +
      "tMapping\0222.policy.subjectmapping.CreateS" +
      "ubjectMappingRequest\0323.policy.subjectmap" +
      "ping.CreateSubjectMappingResponse\"*\202\323\344\223\002" +
      "$\"\021/subject-mappings:\017subject_mapping\022\260\001" +
      "\n\024UpdateSubjectMapping\0222.policy.subjectm" +
      "apping.UpdateSubjectMappingRequest\0323.pol" +
      "icy.subjectmapping.UpdateSubjectMappingR" +
      "esponse\"/\202\323\344\223\002)\"\026/subject-mappings/{id}:" +
      "\017subject_mapping\022\237\001\n\024DeleteSubjectMappin" +
      "g\0222.policy.subjectmapping.DeleteSubjectM" +
      "appingRequest\0323.policy.subjectmapping.De" +
      "leteSubjectMappingResponse\"\036\202\323\344\223\002\030*\026/sub" +
      "ject-mappings/{id}B\364\001\n)io.opentdf.platfo" +
      "rm.policy.subjectmappingB\023SubjectMapping" +
      "ProtoP\001Z=github.com/opentdf/platform/pro" +
      "tocol/go/policy/subjectmapping\242\002\003PSX\252\002\025P" +
      "olicy.Subjectmapping\312\002\025Policy\\Subjectmap" +
      "ping\342\002!Policy\\Subjectmapping\\GPBMetadata" +
      "\352\002\026Policy::Subjectmappingb\006proto3"
    };
    descriptor = com.google.protobuf.Descriptors.FileDescriptor
      .internalBuildGeneratedFileFrom(descriptorData,
        new com.google.protobuf.Descriptors.FileDescriptor[] {
          io.opentdf.platform.authorization.AuthorizationProto.getDescriptor(),
          build.buf.validate.ValidateProto.getDescriptor(),
          io.opentdf.platform.common.CommonProto.getDescriptor(),
          com.google.api.AnnotationsProto.getDescriptor(),
          com.google.protobuf.StructProto.getDescriptor(),
          io.opentdf.platform.policy.attributes.AttributesProto.getDescriptor(),
        });
    internal_static_policy_subjectmapping_Condition_descriptor =
      getDescriptor().getMessageTypes().get(0);
    internal_static_policy_subjectmapping_Condition_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_Condition_descriptor,
        new java.lang.String[] { "SubjectExternalField", "Operator", "SubjectExternalValues", });
    internal_static_policy_subjectmapping_ConditionGroup_descriptor =
      getDescriptor().getMessageTypes().get(1);
    internal_static_policy_subjectmapping_ConditionGroup_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_ConditionGroup_descriptor,
        new java.lang.String[] { "Conditions", "BooleanType", });
    internal_static_policy_subjectmapping_SubjectSet_descriptor =
      getDescriptor().getMessageTypes().get(2);
    internal_static_policy_subjectmapping_SubjectSet_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_SubjectSet_descriptor,
        new java.lang.String[] { "ConditionGroups", });
    internal_static_policy_subjectmapping_SubjectMapping_descriptor =
      getDescriptor().getMessageTypes().get(3);
    internal_static_policy_subjectmapping_SubjectMapping_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_SubjectMapping_descriptor,
        new java.lang.String[] { "Id", "Metadata", "AttributeValue", "SubjectSets", "Actions", });
    internal_static_policy_subjectmapping_Subject_descriptor =
      getDescriptor().getMessageTypes().get(4);
    internal_static_policy_subjectmapping_Subject_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_Subject_descriptor,
        new java.lang.String[] { "Attributes", });
    internal_static_policy_subjectmapping_MatchSubjectMappingsRequest_descriptor =
      getDescriptor().getMessageTypes().get(5);
    internal_static_policy_subjectmapping_MatchSubjectMappingsRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_MatchSubjectMappingsRequest_descriptor,
        new java.lang.String[] { "Subject", });
    internal_static_policy_subjectmapping_MatchSubjectMappingsResponse_descriptor =
      getDescriptor().getMessageTypes().get(6);
    internal_static_policy_subjectmapping_MatchSubjectMappingsResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_MatchSubjectMappingsResponse_descriptor,
        new java.lang.String[] { "SubjectMappings", });
    internal_static_policy_subjectmapping_SubjectMappingCreateUpdate_descriptor =
      getDescriptor().getMessageTypes().get(7);
    internal_static_policy_subjectmapping_SubjectMappingCreateUpdate_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_SubjectMappingCreateUpdate_descriptor,
        new java.lang.String[] { "Metadata", "AttributeValueId", "SubjectSets", "Actions", });
    internal_static_policy_subjectmapping_GetSubjectMappingRequest_descriptor =
      getDescriptor().getMessageTypes().get(8);
    internal_static_policy_subjectmapping_GetSubjectMappingRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_GetSubjectMappingRequest_descriptor,
        new java.lang.String[] { "Id", });
    internal_static_policy_subjectmapping_GetSubjectMappingResponse_descriptor =
      getDescriptor().getMessageTypes().get(9);
    internal_static_policy_subjectmapping_GetSubjectMappingResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_GetSubjectMappingResponse_descriptor,
        new java.lang.String[] { "SubjectMapping", });
    internal_static_policy_subjectmapping_ListSubjectMappingsRequest_descriptor =
      getDescriptor().getMessageTypes().get(10);
    internal_static_policy_subjectmapping_ListSubjectMappingsRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_ListSubjectMappingsRequest_descriptor,
        new java.lang.String[] { });
    internal_static_policy_subjectmapping_ListSubjectMappingsResponse_descriptor =
      getDescriptor().getMessageTypes().get(11);
    internal_static_policy_subjectmapping_ListSubjectMappingsResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_ListSubjectMappingsResponse_descriptor,
        new java.lang.String[] { "SubjectMappings", });
    internal_static_policy_subjectmapping_CreateSubjectMappingRequest_descriptor =
      getDescriptor().getMessageTypes().get(12);
    internal_static_policy_subjectmapping_CreateSubjectMappingRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_CreateSubjectMappingRequest_descriptor,
        new java.lang.String[] { "SubjectMapping", });
    internal_static_policy_subjectmapping_CreateSubjectMappingResponse_descriptor =
      getDescriptor().getMessageTypes().get(13);
    internal_static_policy_subjectmapping_CreateSubjectMappingResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_CreateSubjectMappingResponse_descriptor,
        new java.lang.String[] { "SubjectMapping", });
    internal_static_policy_subjectmapping_UpdateSubjectMappingRequest_descriptor =
      getDescriptor().getMessageTypes().get(14);
    internal_static_policy_subjectmapping_UpdateSubjectMappingRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_UpdateSubjectMappingRequest_descriptor,
        new java.lang.String[] { "Id", "SubjectMapping", });
    internal_static_policy_subjectmapping_UpdateSubjectMappingResponse_descriptor =
      getDescriptor().getMessageTypes().get(15);
    internal_static_policy_subjectmapping_UpdateSubjectMappingResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_UpdateSubjectMappingResponse_descriptor,
        new java.lang.String[] { "SubjectMapping", });
    internal_static_policy_subjectmapping_DeleteSubjectMappingRequest_descriptor =
      getDescriptor().getMessageTypes().get(16);
    internal_static_policy_subjectmapping_DeleteSubjectMappingRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_DeleteSubjectMappingRequest_descriptor,
        new java.lang.String[] { "Id", });
    internal_static_policy_subjectmapping_DeleteSubjectMappingResponse_descriptor =
      getDescriptor().getMessageTypes().get(17);
    internal_static_policy_subjectmapping_DeleteSubjectMappingResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_policy_subjectmapping_DeleteSubjectMappingResponse_descriptor,
        new java.lang.String[] { "SubjectMapping", });
    com.google.protobuf.ExtensionRegistry registry =
        com.google.protobuf.ExtensionRegistry.newInstance();
    registry.add(build.buf.validate.ValidateProto.field);
    registry.add(com.google.api.AnnotationsProto.http);
    com.google.protobuf.Descriptors.FileDescriptor
        .internalUpdateFileDescriptor(descriptor, registry);
    io.opentdf.platform.authorization.AuthorizationProto.getDescriptor();
    build.buf.validate.ValidateProto.getDescriptor();
    io.opentdf.platform.common.CommonProto.getDescriptor();
    com.google.api.AnnotationsProto.getDescriptor();
    com.google.protobuf.StructProto.getDescriptor();
    io.opentdf.platform.policy.attributes.AttributesProto.getDescriptor();
  }

  // @@protoc_insertion_point(outer_class_scope)
}
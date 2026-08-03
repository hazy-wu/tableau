package protogen

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tableauio/tableau/internal/printer"
	"github.com/tableauio/tableau/internal/x/xproto"
	"github.com/tableauio/tableau/options"
	"github.com/tableauio/tableau/proto/tableaupb"
	"github.com/tableauio/tableau/proto/tableaupb/internalpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Test_genFieldOptionsString(t *testing.T) {
	type args struct {
		opts *tableaupb.FieldOptions
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "only-name",
			args: args{
				opts: &tableaupb.FieldOptions{
					Name: "ItemID",
				},
			},
			want: `[(tableau.field) = {name:"ItemID"}]`,
		},
		{
			name: "name-and-prop",
			args: args{
				opts: &tableaupb.FieldOptions{
					Name: "ItemID",
					Prop: &tableaupb.FieldProp{
						Unique: proto.Bool(true),
					},
				},
			},
			want: `[(tableau.field) = {name:"ItemID" prop:{unique:true}}]`,
		},
		{
			name: "name-prop-and-json-name",
			args: args{
				opts: &tableaupb.FieldOptions{
					Name: "ItemID",
					Prop: &tableaupb.FieldProp{
						Unique:   proto.Bool(true),
						JsonName: "item_id_1",
					},
				},
			},
			want: `[(tableau.field) = {name:"ItemID" prop:{unique:true}}, json_name="item_id_1"]`,
		},
		{
			name: "name-and-prop-json_name",
			args: args{
				opts: &tableaupb.FieldOptions{
					Name: "ItemID",
					Prop: &tableaupb.FieldProp{
						JsonName: "item_id_1",
					},
				},
			},
			want: `[(tableau.field) = {name:"ItemID"}, json_name="item_id_1"]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genFieldOptionsString(tt.args.opts); got != tt.want {
				t.Errorf("genFieldOptionsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSameFieldMessageType(t *testing.T) {
	type args struct {
		left  *internalpb.Field
		right *internalpb.Field
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both-are-nil",
			args: args{
				left:  nil,
				right: nil,
			},
			want: true,
		},
		{
			name: "one-is-nil",
			args: args{
				left:  &internalpb.Field{},
				right: nil,
			},
			want: true,
		},
		{
			name: "one-sub-fields-nil",
			args: args{
				left: &internalpb.Field{
					Fields: nil,
				},
				right: &internalpb.Field{
					Fields: []*internalpb.Field{
						{
							Number: 1,
							Name:   "Item",
							Alias:  "RewardItem",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "not-equal-length-of-sub-fields",
			args: args{
				left: &internalpb.Field{
					Fields: []*internalpb.Field{
						{
							Number: 1,
						},
						{
							Number: 2,
						},
					},
				},
				right: &internalpb.Field{
					Fields: []*internalpb.Field{
						{
							Number: 1,
						},
					},
				},
			},
			want: false,
		},
		{
			name: "not-same-type",
			args: args{
				left: &internalpb.Field{
					Type: "Item",
				},
				right: &internalpb.Field{
					Type: "Drop",
				},
			},
			want: false,
		},
		{
			name: "same-sub-fields",
			args: args{
				left: &internalpb.Field{
					Fields: []*internalpb.Field{
						{
							Number: 1,
							Name:   "Item",
							Alias:  "RewardItem",
						},
					},
				},
				right: &internalpb.Field{
					Fields: []*internalpb.Field{
						{
							Number: 1,
							Name:   "Item",
							Alias:  "RewardItem",
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSameFieldMessageType(tt.args.left, tt.args.right); got != tt.want {
				t.Errorf("isSameFieldMessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_marshalToText(t *testing.T) {
	type args struct {
		m protoreflect.ProtoMessage
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "FieldOptions",
			args: args{
				m: &tableaupb.FieldOptions{
					Key:    "ID",
					Layout: tableaupb.Layout_LAYOUT_VERTICAL,
					Prop: &tableaupb.FieldProp{
						Unique:  proto.Bool(true),
						Refer:   "ItemConf.ID",
						Present: true,
					},
				},
			},
			want: `key:"ID" layout:LAYOUT_VERTICAL prop:{unique:true refer:"ItemConf.ID" present:true}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := marshalToText(tt.args.m); got != tt.want {
				t.Errorf("marshalToText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bookExporter_GetProtoFilePath(t *testing.T) {
	tests := []struct {
		name string
		x    *bookExporter
		want string
	}{
		{
			name: "name-and-prop",
			x: &bookExporter{
				wb: &internalpb.Workbook{
					Name: "name",
				},
				FilenameSuffix: "_conf",
			},
			want: `name_conf.proto`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.GetProtoFilePath(); got != tt.want {
				t.Errorf("bookExporter.GetProtoFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sheetExporter_exportEnum(t *testing.T) {
	tests := []struct {
		name    string
		x       *sheetExporter
		want    string
		wantErr bool
	}{
		{
			name: "auto add zero enum value",
			x: &sheetExporter{
				ws: &internalpb.Worksheet{
					Name: "ItemType",
					Options: &tableaupb.WorksheetOptions{
						Name: "ItemType",
					},
					Fields: []*internalpb.Field{
						{Number: 1, Name: "ITEM_TYPE_FRUIT", Alias: "Fruit"},
						{Number: 2, Name: "ITEM_TYPE_EQUIP", Alias: "Equip"},
						{Number: 3, Name: "ITEM_TYPE_BOX", Alias: "Box"},
					},
				},
				p: printer.New(),
				be: &bookExporter{
					gen: &Generator{
						ctx: context.Background(),
					},
				},
			},
			want: `enum ItemType {
  option (tableau.etype) = {name:"ItemType"};

  ITEM_TYPE_INVALID = 0;
  ITEM_TYPE_FRUIT = 1 [(tableau.evalue).name = "Fruit"]; // Fruit
  ITEM_TYPE_EQUIP = 2 [(tableau.evalue).name = "Equip"]; // Equip
  ITEM_TYPE_BOX = 3 [(tableau.evalue).name = "Box"]; // Box
}

`,
			wantErr: false,
		},
		{
			name: "zero enum value not the first one",
			x: &sheetExporter{
				ws: &internalpb.Worksheet{
					Name: "ItemType",
					Options: &tableaupb.WorksheetOptions{
						Name: "ItemType",
					},
					Fields: []*internalpb.Field{
						{Number: -1, Name: "ITEM_TYPE_FRUIT", Alias: "Fruit"},
						{Number: 0, Name: "ITEM_TYPE_EQUIP", Alias: "Equip"},
						{Number: 1, Name: "ITEM_TYPE_BOX", Alias: "Box"},
					},
				},
				p: printer.New(),
				be: &bookExporter{
					gen: &Generator{
						ctx: context.Background(),
					},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.x.exportEnum()
			if (err != nil) != tt.wantErr {
				t.Errorf("sheetExporter.exportEnum() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				assert.Equal(t, tt.want, tt.x.p.String())
			}
		})
	}
}

func Test_sheetExporter_exportStruct(t *testing.T) {
	tests := []struct {
		name    string
		x       *sheetExporter
		want    string
		wantErr bool
	}{
		{
			name: "export-struct",
			x: &sheetExporter{
				ws: &internalpb.Worksheet{
					Name: "TaskReward",
					Options: &tableaupb.WorksheetOptions{
						Name: "StructTaskReward",
					},
					Fields: []*internalpb.Field{
						{Name: "id", Type: "uint32", FullType: "uint32", Options: &tableaupb.FieldOptions{Name: "ID"}},
						{Name: "num", Type: "int32", FullType: "int32", Options: &tableaupb.FieldOptions{Name: "Num"}},
						{Name: "fruit_type", Type: "FruitType", FullType: "protoconf.FruitType", Predefined: true, Options: &tableaupb.FieldOptions{Name: "FruitType"}},
					},
				},
				p:              printer.New(),
				typeInfos:      &xproto.TypeInfos{},
				nestedMessages: make(map[string]*internalpb.Field),
			},
			want: `message TaskReward {
  option (tableau.struct) = {name:"StructTaskReward"};

  uint32 id = 1 [(tableau.field) = {name:"ID"}];
  int32 num = 2 [(tableau.field) = {name:"Num"}];
  protoconf.FruitType fruit_type = 3 [(tableau.field) = {name:"FruitType"}];
}

`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.exportStruct(); (err != nil) != tt.wantErr {
				t.Errorf("sheetExporter.exportStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.x.p.String())
		})
	}
}

func Test_sheetExporter_exportUnion(t *testing.T) {
	// auxWant describes an expected shard file emitted by union splitting.
	type auxWant struct {
		relPath string
		body    string
	}
	tests := []struct {
		name    string
		x       *sheetExporter
		want    string
		wantAux []auxWant
		wantErr bool
	}{
		{
			name: "export-union",
			x: &sheetExporter{
				ws: &internalpb.Worksheet{
					Name: "TaskTarget",
					Options: &tableaupb.WorksheetOptions{
						Name: "UnionTaskTarget",
					},
					Fields: []*internalpb.Field{
						{Number: 1, Name: "PvpBattle", Alias: "SoloPVPBattle",
							Fields: []*internalpb.Field{
								{Number: 1, Name: "id", Type: "uint32", FullType: "uint32", Options: &tableaupb.FieldOptions{Name: "ID"}},
								{Number: 2, Name: "damage", Type: "int64", FullType: "int64", Options: &tableaupb.FieldOptions{Name: "Damage"}},
								{Number: 3, Name: "type_list", Type: "repeated", FullType: "repeated protoconf.FruitType",
									ListEntry:  &internalpb.Field_ListEntry{ElemType: "FruitType", ElemFullType: "protoconf.FruitType"},
									Predefined: true,
									Options:    &tableaupb.FieldOptions{Name: "Type", Layout: tableaupb.Layout_LAYOUT_INCELL}},
							},
						},
					},
				},
				p: printer.New(),
				be: &bookExporter{
					gen: &Generator{
						ctx: context.Background(),
					},
				},
				typeInfos:      &xproto.TypeInfos{},
				nestedMessages: make(map[string]*internalpb.Field),
			},
			want: `message TaskTarget {
  option (tableau.union) = {name:"UnionTaskTarget"};

  Type type = 9999 [(tableau.field) = {name:"Type"}];
  oneof value {
    option (tableau.oneof) = {field:"Field"};

    PvpBattle pvp_battle = 1; // Bound to enum value: TYPE_PVP_BATTLE.
  }

  enum Type {
    TYPE_INVALID = 0;
    TYPE_PVP_BATTLE = 1 [(tableau.evalue).name = "SoloPVPBattle"]; // SoloPVPBattle
  }

  message PvpBattle {
    uint32 id = 1 [(tableau.field) = {name:"ID"}];
    int64 damage = 2 [(tableau.field) = {name:"Damage"}];
    repeated protoconf.FruitType type_list = 3 [(tableau.field) = {name:"Type" layout:LAYOUT_INCELL}];
  }
}

`,
			wantErr: false,
		},
		{
			// export-union-split verifies that when InputOpt.UnionSplitThreshold
			// is set and exceeded, sub-messages are extracted into shard files
			// as top-level messages with the "T_" separator, and the main
			// union body no longer emits nested message blocks.
			name: "export-union-split",
			x: &sheetExporter{
				ws: &internalpb.Worksheet{
					Name: "TaskTarget",
					Options: &tableaupb.WorksheetOptions{
						Name: "UnionTaskTarget",
					},
					Fields: []*internalpb.Field{
						{Number: 1, Name: "PvpBattle", Alias: "SoloPVPBattle",
							Fields: []*internalpb.Field{
								{Number: 1, Name: "id", Type: "uint32", FullType: "uint32", Options: &tableaupb.FieldOptions{Name: "ID"}},
							},
						},
						{Number: 2, Name: "PveBattle", Alias: "SoloPVEBattle",
							Fields: []*internalpb.Field{
								{Number: 1, Name: "level", Type: "uint32", FullType: "uint32", Options: &tableaupb.FieldOptions{Name: "Level"}},
							},
						},
					},
				},
				p: printer.New(),
				be: &bookExporter{
					FilenameSuffix: "",
					gen: &Generator{
						ctx: context.Background(),
						InputOpt: &options.ProtoInputOption{
							// Threshold=1 with 2 sub-messages triggers split.
							// ShardSize=1 puts each sub-message in its own shard,
							// yielding exactly 2 shards for deterministic asserts.
							UnionSplitThreshold: 1,
							UnionSplitShardSize: 1,
						},
					},
					// wb.Name drives GetProtoFilePath(); we use "task" so the
					// shards are named task_task_target_{1,2}.proto.
					wb: &internalpb.Workbook{Name: "task"},
				},
				typeInfos:      &xproto.TypeInfos{},
				nestedMessages: make(map[string]*internalpb.Field),
			},
			want: `message TaskTarget {
  option (tableau.union) = {name:"UnionTaskTarget"};

  Type type = 9999 [(tableau.field) = {name:"Type"}];
  oneof value {
    option (tableau.oneof) = {field:"Field"};

    TaskTargetT_PvpBattle pvp_battle = 1; // Bound to enum value: TYPE_PVP_BATTLE.
    TaskTargetT_PveBattle pve_battle = 2; // Bound to enum value: TYPE_PVE_BATTLE.
  }

  enum Type {
    TYPE_INVALID = 0;
    TYPE_PVP_BATTLE = 1 [(tableau.evalue).name = "SoloPVPBattle"]; // SoloPVPBattle
    TYPE_PVE_BATTLE = 2 [(tableau.evalue).name = "SoloPVEBattle"]; // SoloPVEBattle
  }
}

`,
			wantAux: []auxWant{
				{
					relPath: "task_task_target_1.proto",
					body: `message TaskTargetT_PvpBattle {
  uint32 id = 1 [(tableau.field) = {name:"ID"}];
}
`,
				},
				{
					relPath: "task_task_target_2.proto",
					body: `message TaskTargetT_PveBattle {
  uint32 level = 1 [(tableau.field) = {name:"Level"}];
}
`,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.exportUnion(); (err != nil) != tt.wantErr {
				t.Errorf("sheetExporter.exportUnion() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.x.p.String())
			// Verify shard files if the test case expects splitting.
			if len(tt.wantAux) > 0 {
				assert.Equal(t, len(tt.wantAux), len(tt.x.be.auxFiles),
					"unexpected number of aux shard files")
				assert.Equal(t, len(tt.wantAux), len(tt.x.be.auxImportPaths),
					"aux shard count mismatch with import paths")
				for i, want := range tt.wantAux {
					if i >= len(tt.x.be.auxFiles) {
						break
					}
					got := tt.x.be.auxFiles[i]
					assert.Equal(t, want.relPath, got.RelPath,
						"aux[%d] RelPath mismatch", i)
					assert.Equal(t, want.body, got.Printer.String(),
						"aux[%d] body mismatch", i)
					assert.Equal(t, want.relPath, tt.x.be.auxImportPaths[i],
						"aux[%d] import path mismatch", i)
				}
			} else {
				assert.Empty(t, tt.x.be.auxFiles,
					"unexpected aux files for non-split case")
			}
		})
	}
}

func Test_sheetExporter_exportMessager(t *testing.T) {
	tests := []struct {
		name    string
		x       *sheetExporter
		want    string
		wantErr bool
	}{
		{
			name: "export-messager",
			x: &sheetExporter{
				ws: &internalpb.Worksheet{
					Name: "ItemConf",
					Fields: []*internalpb.Field{
						{Name: "id", Type: "uint32", FullType: "uint32", Options: &tableaupb.FieldOptions{Name: "ID"}},
					},
				},
				p: printer.New(),
				be: &bookExporter{
					gen:                   &Generator{},
					messagerPatternRegexp: regexp.MustCompile(`Conf$`),
				},
				typeInfos:      &xproto.TypeInfos{},
				nestedMessages: make(map[string]*internalpb.Field),
			},
			want: `message ItemConf {
  option (tableau.worksheet) = {};

  uint32 id = 1 [(tableau.field) = {name:"ID"}];
}

`,
			wantErr: false,
		},
		{
			name: "export-messager-pattern-not-match",
			x: &sheetExporter{
				ws: &internalpb.Worksheet{
					Name: "ItemConf",
					Fields: []*internalpb.Field{
						{Name: "id", Type: "uint32", FullType: "uint32", Options: &tableaupb.FieldOptions{Name: "ID"}},
					},
				},
				p: printer.New(),
				be: &bookExporter{
					gen:                   &Generator{},
					messagerPatternRegexp: regexp.MustCompile(`Data$`),
				},
				typeInfos:      &xproto.TypeInfos{},
				nestedMessages: make(map[string]*internalpb.Field),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.x.exportMessager(); (err != nil) != tt.wantErr {
				t.Errorf("sheetExporter.exportMessager() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.x.p.String())
		})
	}
}

package metrics

import (
	"fmt"
	"reflect"
	"strings"
)

// Init initializes the passed in metrics and initializes its fields using the passed in factory.
func Init(metrics interface{}, factory Factory, globalTags map[string]string) {
	if err := initMetrics(metrics, factory, globalTags); err != nil {
		panic(err.Error())
	}
}

// initMetrics uses reflection to initialize a struct containing metrics fields
// by assigning new Counter/Gauge/Timer values with the metric name retrieved
// from the `metric` tag and stats tags retrieved from the `tags` tag.
//
// Note: all fields of the struct must be exported, have a `metric` tag, and be
// of type Counter or Gauge or Timer.
func initMetrics(m interface{}, factory Factory, globalTags map[string]string) error {
	// Allow user to opt out of reporting metrics by passing in nil.
	if factory == nil {
		factory = NullFactory
	}

	counterPtrType := reflect.TypeOf((*Counter)(nil)).Elem()
	gaugePtrType := reflect.TypeOf((*Gauge)(nil)).Elem()
	timerPtrType := reflect.TypeOf((*Timer)(nil)).Elem()

	v := reflect.ValueOf(m).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		tags := make(map[string]string)
		for k, v := range globalTags {
			tags[k] = v
		}
		field := t.Field(i)
		metric := field.Tag.Get("metric")
		if metric == "" {
			return fmt.Errorf("Field %s is missing a tag 'metric'", field.Name)
		}
		if tagString := field.Tag.Get("tags"); tagString != "" {
			tagPairs := strings.Split(tagString, ",")
			for _, tagPair := range tagPairs {
				tag := strings.Split(tagPair, "=")
				if len(tag) != 2 {
					return fmt.Errorf(
						"Field [%s]: Tag [%s] is not of the form key=value in 'tags' string [%s]",
						field.Name, tagPair, tagString)
				}
				tags[tag[0]] = tag[1]
			}
		}
		var obj interface{}
		if field.Type.AssignableTo(counterPtrType) {
			obj = factory.CreateCounter(metric, tags)
		} else if field.Type.AssignableTo(gaugePtrType) {
			obj = factory.CreateGauge(metric, tags)
		} else if field.Type.AssignableTo(timerPtrType) {
			obj = factory.CreateTimer(metric, tags)
		} else {
			return fmt.Errorf(
				"Field %s is not a pointer to timer, gauge, or counter",
				field.Name)
		}
		v.Field(i).Set(reflect.ValueOf(obj))
	}
	return nil
}

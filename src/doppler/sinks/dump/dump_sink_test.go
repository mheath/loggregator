package dump_test

import (
	"doppler/sinks/dump"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/stretchr/testify/assert"
	"runtime"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Dump Sink", func() {
	It("works with one message", func() {

		testDump := dump.NewDumpSink("myApp", 1, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "hi", "appId", "App"), "origin")
		inputChan <- logMessage

		close(inputChan)
		<-dumpRunnerDone

		data := testDump.Dump()
		assert.Equal(GinkgoT(), len(data), 1)
		Expect(string(data[0].GetLogMessage().GetMessage())).To(Equal("hi"))
	})

	It("works with two messages", func() {

		testDump := dump.NewDumpSink("myApp", 2, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "1", "appId", "App"), "origin")
		inputChan <- logMessage
		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "2", "appId", "App"), "origin")
		inputChan <- logMessage

		close(inputChan)
		<-dumpRunnerDone

		logMessages := testDump.Dump()

		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("1"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("2"))
	})

	It("never fills up", func() {

		bufferSize := uint32(3)
		testDump := dump.NewDumpSink("myApp", bufferSize, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "hi", "appId", "App"), "origin")

		for i := uint32(0); i < bufferSize+1; i++ {
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone
	})

	It("always returns the newest messages", func() {

		testDump := dump.NewDumpSink("myApp", 2, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})

		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "1", "appId", "App"), "origin")
		inputChan <- logMessage
		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "2", "appId", "App"), "origin")
		inputChan <- logMessage
		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "3", "appId", "App"), "origin")
		inputChan <- logMessage

		close(inputChan)
		<-dumpRunnerDone

		logMessages := testDump.Dump()
		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("2"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("3"))
	})

	It("returns all recent messages to multiple dump requests", func() {

		testDump := dump.NewDumpSink("myApp", 2, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "1", "appId", "App"), "origin")
		inputChan <- logMessage
		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "2", "appId", "App"), "origin")
		inputChan <- logMessage
		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "3", "appId", "App"), "origin")
		inputChan <- logMessage

		close(inputChan)
		<-dumpRunnerDone

		logMessages := testDump.Dump()
		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("2"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("3"))

		logMessages = testDump.Dump()
		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("2"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("3"))
	})

	It("returns all recent messages to multiple dump requests with messages cloning in in the meantime", func() {

		testDump := dump.NewDumpSink("myApp", 2, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "1", "appId", "App"), "origin")
		inputChan <- logMessage
		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "2", "appId", "App"), "origin")
		inputChan <- logMessage
		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "3", "appId", "App"), "origin")
		inputChan <- logMessage

		close(inputChan)
		<-dumpRunnerDone

		logMessages := testDump.Dump()
		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("2"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("3"))

		dumpRunnerDone = make(chan struct{})
		inputChan = make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		logMessage, _ = emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, "4", "appId", "App"), "origin")
		inputChan <- logMessage

		Eventually(func() string {
			logMessages = testDump.Dump()
			return string(logMessages[0].GetLogMessage().GetMessage())
		}).Should(Equal("3"))

		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("4"))
	})

	It("works with lots of messages", func() {

		testDump := dump.NewDumpSink("myApp", 2, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 0; i < 100; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone

		logMessages := testDump.Dump()
		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("98"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("99"))

		dumpRunnerDone = make(chan struct{})
		inputChan = make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 100; i < 200; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone

		logMessages = testDump.Dump()
		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("198"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("199"))

		logMessages = testDump.Dump()
		Expect(logMessages).To(HaveLen(2))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("198"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("199"))
	})

	It("works with lots of messages and large buffer", func() {

		testDump := dump.NewDumpSink("myApp", 200, loggertesthelper.Logger(), time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 0; i < 1000; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone

		logMessages := testDump.Dump()
		Expect(logMessages).To(HaveLen(200))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("800"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("801"))

		dumpRunnerDone = make(chan struct{})
		inputChan = make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 1000; i < 2000; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone

		logMessages = testDump.Dump()
		Expect(logMessages).To(HaveLen(200))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("1800"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("1801"))

		logMessages = testDump.Dump()
		Expect(logMessages).To(HaveLen(200))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("1800"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("1801"))
	})

	It("works with lots of messages and large buffer2", func() {

		testDump := dump.NewDumpSink("myApp", 200, loggertesthelper.Logger(), time.Second)
		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 0; i < 100; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone

		logMessages := testDump.Dump()
		Expect(logMessages).To(HaveLen(100))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("0"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("1"))
		Expect(string(logMessages[99].GetLogMessage().GetMessage())).To(Equal("99"))

		dumpRunnerDone = make(chan struct{})
		inputChan = make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 100; i < 200; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone
		logMessages = testDump.Dump()
		Expect(logMessages).To(HaveLen(200))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("0"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("1"))

		dumpRunnerDone = make(chan struct{})
		inputChan = make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 200; i < 300; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone
		logMessages = testDump.Dump()
		Expect(logMessages).To(HaveLen(200))
		Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("100"))
		Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("101"))
	})

	It("works with lots of dumps", func() {

		runtime.GOMAXPROCS(runtime.NumCPU())
		testDump := dump.NewDumpSink("myApp", 5, loggertesthelper.Logger(), time.Second)
		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		for i := 0; i < 10; i++ {
			logMessage, _ := emitter.Wrap(factories.NewLogMessage(events.LogMessage_OUT, strconv.Itoa(i), "appId", "App"), "origin")
			inputChan <- logMessage
		}

		close(inputChan)
		<-dumpRunnerDone

		for i := 0; i < 200; i++ {
			go func() {
				logMessages := testDump.Dump()

				Expect(logMessages).To(HaveLen(5))
				Expect(string(logMessages[0].GetLogMessage().GetMessage())).To(Equal("5"))
				Expect(string(logMessages[1].GetLogMessage().GetMessage())).To(Equal("6"))
			}()
		}
	})

	It("closes itself after period of inactivity", func() {

		testDump := dump.NewDumpSink("myApp", 5, loggertesthelper.Logger(), 2*time.Microsecond)
		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		select {
		case <-dumpRunnerDone:

		case <-time.After(200 * time.Millisecond):
			assert.Fail(GinkgoT(), "Should have timed out the dump")
		}
	})

	It("only stores log messages", func() {
		testDump := dump.NewDumpSink("myApp", 5, loggertesthelper.Logger(), 2*time.Second)

		dumpRunnerDone := make(chan struct{})
		inputChan := make(chan *events.Envelope, 5)

		go func() {
			testDump.Run(inputChan)
			close(dumpRunnerDone)
		}()

		var env *events.Envelope
		env, _ = emitter.Wrap(&events.Heartbeat{}, "origin")
		inputChan <- env
		env, _ = emitter.Wrap(&events.LogMessage{}, "origin") // should keep this one
		inputChan <- env
		env, _ = emitter.Wrap(&events.HttpStartStop{}, "origin")
		inputChan <- env
		env, _ = emitter.Wrap(&events.ValueMetric{}, "origin")
		inputChan <- env

		close(inputChan)
		<-dumpRunnerDone

		Expect(testDump.Dump()).To(HaveLen(1))
	})
})
